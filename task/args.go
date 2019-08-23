package task

import (
	"../crc"
	"../dts"
	"../etcd"
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	version         = "0.5"
	dtsAppName      = "GO-DTS"
	dtsDefaultDir   = "/data/usnmp/go-dts"
	dtsCustomDir    = "C:/gitdir"
	etcdDefaultPort = "2500"
	etcdCustomPort  = "2379"
	etcdTestUrlPref = "vlg-mon-app1"
	etcGreenUrl     = "influx.megafon.ru"
	etcdCustomUrl   = "centos.emink.net"
	dtsId           = "5118"
	logDir          = "/data/logs/go-dts/"
	//logCustomDir    = "C:/gitdir/logs/"
)

var (
	parser                 *flags.Parser
	ErrInstanceNotExist    = errors.New("instance do not exists in etcd")
	ErrDtsAppExtracting    = errors.New("unable to extract dts app by specified instance")
	ErrTargetAppExtracting = errors.New("unable to extract target app by specified instance")
	ErrInstanceExist       = errors.New("instance already exists")
	ErrAppDirNotMatch      = errors.New("app dirs do not match")
	ErrWorkTreeNotMatch    = errors.New("work tree's do not matches")
	ErrAppNameNotMatch     = errors.New("app names not matches")
	ErrInstanceDisabled    = errors.New("instance disabled")
	ErrInstancesNotMatch   = errors.New("instances do not match")
)

func init() {
	setupLogger()
}

// Init new parser
func (st *State) NewParser() {
	st.Opts = &Options{}
	parser = flags.NewParser(st.Opts, flags.Default)
	parser.Usage = "--action=[init,status] [--work-tree [--dts-dir], --instance]"
}

// Parse arguments
func (st *State) ParseArgs() {
	_, err := parser.ParseArgs(os.Args)
	st.checkError(err)
}

// Fetch data from everywhere
func (st *State) Fetch() {
	st.helpOrVersion()

	err := st.prepare()
	st.checkError(err)

	if st.Opts.Test {
		Log.Printf("args before: %+v\n", *st.Vars)

		st.customArgs()
		Log.Printf("args after: %+v\n", *st.Vars)
	}

	st.config = &etcd.Etcd{}

	err = st.config.FetchConfig(st.Vars.EtcdUrl, st.Vars.Hostname)
	st.checkError(err)

	st.DtsApp = &etcd.App{}
	ok, err := st.config.FetchAppByInstance(st.Vars.DtsInstance, st.DtsApp)
	st.checkError(err)

	if !ok {
		if st.Opts.Action == "status" {
			st.checkError(ErrDtsAppExtracting)
		}

		st.DtsApp.DtsSettings = &etcd.DtsSettings{}
		st.DtsApp.EmonJson = &etcd.EmonJson{}
	}
}

// Prepare vars
func (st *State) prepare() (err error) {
	vars := &Vars{}
	switch st.Opts.Action {
	case "init":
		if len(st.Opts.WorkTree) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "work-tree required for init action"}
			break
		}

		vars.CurrentLess = leaveCurrent(st.Opts.WorkTree)
		st.Opts.Instance = getInstance(vars.CurrentLess)
		Log.Println("current less:", vars.CurrentLess, "instance:", st.Opts.Instance)
	case "status":
		if len(st.Opts.Instance) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "instance required for status action"}
		}
		//default:
		//	err = &flags.Error{Type: flags.ErrInvalidChoice, Message: "invalid action value, should be [init, status]"}
	}

	if err != nil {
		return
	}

	if len(st.Opts.DtsDir) == 0 {
		st.Opts.DtsDir = dtsDefaultDir
	}

	vars.DtsInstance = getInstance(st.Opts.DtsDir)

	vars.Hostname, err = getShortHostName()
	if err != nil {
		return
	}

	vars.EtcdUrl = getEtcdUrl(vars.Hostname)

	st.Vars = vars

	return
}

func (st *State) Deploy(ex []string) {
	apps := st.config.CollectApps(ex)

	st.checkError(etcd.SetEtcdApi(st.Vars.EtcdUrl))

	for i := 0; i < len(apps); i++ {
		st.Opts.Instance = apps[i][1]
		st.TApp = &etcd.App{}
		ok, err := st.config.FetchAppByInstance(st.Opts.Instance, st.TApp)
		st.checkError(err)

		if !ok {
			st.checkError(ErrTargetAppExtracting)
		}

		if getInstance(st.TApp.AppDir) != st.Opts.Instance {
			st.checkError(ErrInstancesNotMatch)
		}

		st.Opts.WorkTree = st.TApp.AppDir

		if _, ok = st.DtsApp.DtsSettings.AppList[st.Opts.Instance]; ok {
			st.checkError(ErrInstanceExist)
		}

		st.Vars.CurrentLess = st.TApp.AppDir

		st.setDtsApp()

		st.checkError(st.gitInit())

		st.checkError(st.logJson())
	}

	city := strings.Split(st.Vars.Hostname, "-")[0]
	uri := fmt.Sprintf("/ps/hosts/%s/%s/apps/%s.%s/", city, st.Vars.Hostname, dtsId, st.Vars.DtsInstance)
	updatedKeys, err := st.DtsApp.Push(uri)
	st.checkError(err)

	Log.Println(updatedKeys)
}

func (st *State) Init() {
	st.TApp = &etcd.App{}
	ok, err := st.config.FetchAppByInstance(st.Opts.Instance, st.TApp)
	st.checkError(err)

	if !ok {
		st.checkError(ErrTargetAppExtracting)
	}

	if _, ok = st.DtsApp.DtsSettings.AppList[st.Opts.Instance]; ok {
		st.checkError(ErrInstanceExist)
	}

	if len(st.Vars.CurrentLess) != 0 && st.TApp.AppDir != st.Vars.CurrentLess {
		st.checkError(ErrAppDirNotMatch)
	}

	st.setDtsApp()

	err = st.gitInit()
	st.checkError(err)

	err = etcd.SetEtcdApi(st.Vars.EtcdUrl)
	st.checkError(err)

	uri := fmt.Sprintf("/ps/hosts/%s/apps/%s.%s/", st.Vars.Hostname, dtsId, st.Vars.DtsInstance)
	updateKeys, err := st.DtsApp.Push(uri)
	st.checkError(err)

	Log.Println(updateKeys)
}

// Init external git dir and add accessible files
func (st *State) gitInit() error {
	Log.Println(st.Opts.WorkTree, st.Opts.DtsDir, st.Opts.Instance)
	wt, err := dts.Init(st.Opts.WorkTree, st.Opts.DtsDir, st.Opts.Instance)
	if err != nil {
		return err
	}

	st.Files = &Files{}
	err = st.Files.walk(st.Opts.WorkTree)
	if err != nil {
		return err
	}

	Log.Printf("accessible: %v, unReadable: %v, gtSize: %v, symlinks: %v\n", st.Files.Accessible, st.Files.UnReadable, st.Files.GtSize, st.Files.Symlinks)
	// Add & commit
	committed, err := dts.Commit(st.Files.Accessible, wt)
	if err != nil {
		return err
	}

	Log.Println("committed:", committed)

	return err
}

func (st *State) Status() {
	st.TApp = &etcd.App{}
	ok, err := st.config.FetchAppByInstance(st.Opts.Instance, st.TApp)
	st.checkError(err)

	if !ok {
		st.checkError(ErrTargetAppExtracting)
	}

	// Check dts config with target application
	v, ok := st.DtsApp.DtsSettings.AppList[st.Opts.Instance]
	if ok {
		if !st.DtsApp.DtsSettings.AppList[st.Opts.Instance].Enabled {
			st.checkError(ErrInstanceDisabled)
		}

		st.Vars.CurrentLess = leaveCurrent(v.WorkTree)
		if st.TApp.AppDir != st.Vars.CurrentLess {
			st.checkError(ErrWorkTreeNotMatch)
		}

		if st.TApp.ApplicationName != v.AppName {
			st.checkError(ErrAppNameNotMatch)
		}

		repo, err := dts.Open(v.WorkTree, v.GitDir)
		st.checkError(err)

		var diffs dts.MFiles
		diffs, err = dts.Diff(repo)
		st.checkError(err)

		st.MFiles = &diffs
	} else {
		st.checkError(ErrInstanceNotExist)
	}

	return
}

// Telegraf stdout string in telegraf format
func (st *State) Telegraf() {
	str := st.MFiles.Telegraf(st.DtsApp.DtsSettings.AppList[st.Opts.Instance].AppName)
	if len(str) != 0 {
		Log.Println(str)
		fmt.Println(str)
	}
}

// LogJson marshal current state into json format and write it into default json log
func (st *State) LogJson() {
	st.checkError(st.logJson())
}

// Assignee custom args if required flag was specified
func (st *State) customArgs() {
	if st.Opts.Test {
		st.Vars.Hostname = "vlg-lbrt-app1d"
		st.Vars.EtcdUrl = fmt.Sprintf("http://%s:%s", etcdCustomUrl, etcdCustomPort)
		st.Opts.DtsDir = dtsCustomDir
		st.Vars.DtsInstance = getInstance(dtsCustomDir)
		if len(st.Opts.WorkTree) != 0 {
			st.Vars.CurrentLess = st.Opts.WorkTree
		}
	}
}

// Update dts app struct, combine next function (SetDtsSettings, SetEmonJson, SetDtsApp)
func (st *State) setDtsApp() {
	st.DtsApp.DtsSettings.SetDtsSettings(st.TApp.ApplicationName, st.Opts.WorkTree, st.Opts.DtsDir, st.Opts.Instance)
	st.DtsApp.EmonJson.SetEmonJson(dtsId, dtsAppName, st.Opts.DtsDir, st.Opts.Instance)
	st.DtsApp.SetDtsApp(dtsId, dtsAppName, st.TApp.Stand, st.DtsApp.DtsSettings, st.DtsApp.EmonJson)
}

// Print help or version if required flags were specified
func (st *State) helpOrVersion() {
	if st.Opts.Help.Help {
		//logFile.Close()
		parser.WriteHelp(os.Stderr)
		os.Exit(0)
	}

	if st.Opts.Help.Version {
		//logFile.Close()
		log.Printf("%s: v%s", dtsAppName, version)
		os.Exit(0)
	}
}

// Get etcd url based on short host name
func getEtcdUrl(sName string) string {
	var url string
	testZone := regexp.MustCompile(`^([a-z]{2,4}-?){3}\d+[a-z]$`)

	if testZone.Match([]byte(sName)) {
		url = fmt.Sprintf("http://%s%s:%s", etcdTestUrlPref, sName[len(sName)-1:], etcdDefaultPort)
	} else {
		url = fmt.Sprintf("http://%s:%s", etcGreenUrl, etcdDefaultPort)
	}

	return url
}

func getInstance(workTree string) string {
	return strconv.FormatUint(uint64(crc.CkSum(workTree+"\n")), 10)
}

func (st *State) checkError(err error) {
	if err != nil {
		st.Err = err.Error()
		if err := st.logJson(); err != nil {
			Log.Println("can't write json log:", err)
		}

		Log.Panic(err)
	}
}
