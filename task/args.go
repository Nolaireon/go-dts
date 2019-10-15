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
	version         = "0.7"
	dtsAppName      = "GO-DTS"
	etcdDefaultPort = "2500"
	etcdTestUrlPref = "vlg-mon-app1"
	etcdGreenUrl    = "influx.megafon.ru"
	dtsApplId       = 5118
	logDir          = "/data/logs/go-dts"
)

var (
	rmEscape               = strings.NewReplacer("\\n", "\n", "\\t", "\t", "\\r\\n", "\r\n")
	errInstanceIsNotExist  = errors.New("instance do not exists in etcd")
	errExtractingDtsApp    = errors.New("unable to extract dts app by specified instance")
	errExtractingTargetApp = errors.New("unable to extract target app by specified instance")
	errInstanceIsExist     = errors.New("instance already exists")
	errAppDirNotMatch      = errors.New("app dirs do not match")
	errAppNameNotMatch     = errors.New("app names not matches")
	errInstanceDisabled    = errors.New("instance disabled")
	errInstancesNotMatch   = errors.New("instances do not match")
	//ErrWorkTreeNotMatch    = errors.New("work tree's do not matches")
)

func init() {
	setupLogger()
}

// ParseArgs parse command-line arguments to the given structure
func (st *State) ParseArgs(args []string) {
	st.Args = &Arguments{}
	parser := flags.NewParser(st.Args, flags.Default)
	parser.Usage = "--action=[init,status,deploy] [--work-tree [--dts-dir], --instance]"
	if len(args) == 0 {
		args = os.Args
	}
	_, err := parser.ParseArgs(args)
	st.checkError(err)

	err = st.checkArgs(parser)
	st.checkError(err)
}

// Check parsed arguments
func (st *State) checkArgs(parser *flags.Parser) (err error) {
	// Handling help group arguments (help, version)
	//st.helpOrVersion(parser)

	if st.Args.Help.Help {
		parser.WriteHelp(os.Stderr)
		os.Exit(0)
	}

	if st.Args.Help.Version {
		log.Printf("%s: v%s", dtsAppName, version)
		os.Exit(0)
	}

	switch st.Args.Action {
	case "init":
		if len(st.Args.WorkTree) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "work-tree required for init action"}
		}
	case "status":
		if len(st.Args.Instance) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "instance required for status action"}
		}
	}

	return
}

// PrepareEnv populate Environment structure before main task execution
func (st *State) PrepareEnv() {
	var err error
	env := &Environment{}

	if st.Args.Action == "init" {
		err = env.decomposeWorkTree(st.Args.WorkTree)
		st.checkError(err)
	}

	if st.Args.Action == "status" {
		env.Instance = st.Args.Instance
	}

	env.DtsDir, err = getExecutablePath()
	st.checkError(err)

	env.DtsInstance = getInstance(env.DtsDir)

	env.Hostname, err = getShortHostName()
	st.checkError(err)

	env.EtcdUrl = getEtcdUrl(env.Hostname)

	st.Env = env
	Log.Printf("env: %+v\n", *env)

	if st.Args.Test {
		configPath := joinPaths("config", "custom_env.yml")
		Log.Println("config path:", configPath)
		err = st.replaceEnv(configPath)
		if err != nil {
			Log.Printf("can't read %s: %s\n", configPath, err)
		} else {
			Log.Printf("modified env: %+v\n", *st.Env)
		}
	}
}

// Fetch data from registry host
func (st *State) Fetch() {
	st.config = &etcd.Etcd{}

	city := strings.Split(st.Env.Hostname, "-")[0]
	// Remove below condition after tests
	if st.Args.Test {
		city = "test"
	}

	url := fmt.Sprintf("%s/v2/keys/ps/hosts/%s/%s/apps?recursive=true", st.Env.EtcdUrl, city, st.Env.Hostname)

	err := st.config.FetchConfig(url)
	st.checkError(err)

	st.DtsApp = &etcd.App{}
	ok, err := st.config.FetchAppByInstance(st.Env.DtsInstance, st.DtsApp)
	st.checkError(err)

	if !ok {
		if st.Args.Action == "status" {
			st.checkError(errExtractingDtsApp)
		}

		st.DtsApp.DtsSettings = &etcd.DtsSettings{}
		st.DtsApp.EmonJson = &etcd.EmonJson{}
	}
}

func (st *State) Deploy() {
	configPath := joinPaths("config", "excluded_apps.yml")
	ea, err := getExcludedApps(configPath)
	if err != nil {
		Log.Printf("can't read %s: %s\n", configPath, err)
	} else {
		Log.Println("parsed list of excluded apps from the config:", configPath)
	}

	apps := st.config.CollectApps(ea)

	Log.Println("deploy apps:", apps)
	err = etcd.SetEtcdApi(st.Env.EtcdUrl)
	st.checkError(err)

	for i := 0; i < len(apps); i++ {
		st.Env.Instance = apps[i][1]
		st.TApp = &etcd.App{}
		var ok bool
		ok, err = st.config.FetchAppByInstance(st.Env.Instance, st.TApp)
		st.checkError(err)

		if !ok {
			st.checkError(errExtractingTargetApp)
		}

		if getInstance(st.TApp.AppDir) != st.Env.Instance {
			Log.Println(errInstancesNotMatch, "continue...")
			//st.checkError(errInstancesNotMatch)
		}

		st.Env.AppDir = st.TApp.AppDir
		st.Env.WorkTree, err = resolveCurrentVersion(st.Env.AppDir)
		st.checkError(err)

		if _, ok = st.DtsApp.DtsSettings.AppList[st.Env.Instance]; ok {
			Log.Printf("Instance '%s' is already exists. Continue...", st.Env.Instance)
			continue
			//st.checkError(errInstanceIsExist)
		}

		st.setDtsApp()

		st.checkError(st.gitInit())

		st.checkError(st.logJson())
	}

	city := strings.Split(st.Env.Hostname, "-")[0]
	if st.Args.Test {
		city = "test"
	}

	uri := fmt.Sprintf("/ps/hosts/%s/%s/apps/%d.%s/", city, st.Env.Hostname, dtsApplId, st.Env.DtsInstance)
	updatedKeys, err := st.DtsApp.Push(uri)
	st.checkError(err)

	Log.Println(updatedKeys)
}

func (st *State) PlainInit() {
	st.TApp = &etcd.App{}
	ok, err := st.config.FetchAppByInstance(st.Env.Instance, st.TApp)
	st.checkError(err)

	if !ok {
		st.checkError(errExtractingTargetApp)
	}

	if _, ok = st.DtsApp.DtsSettings.AppList[st.Env.Instance]; ok {
		st.checkError(errInstanceIsExist)
	}

	if st.TApp.AppDir != st.Env.AppDir {
		st.checkError(errAppDirNotMatch)
	}

	st.checkError(st.init())
}

// init function of State can be calling alone in case when steps described in PlainInit function were done somewhere else
func (st *State) init() (err error) {
	st.setDtsApp()

	err = st.gitInit()
	if err != nil {
		return
	}

	err = etcd.SetEtcdApi(st.Env.EtcdUrl)
	if err != nil {
		return
	}

	city := strings.Split(st.Env.Hostname, "-")[0]
	// Remove below condition after tests
	if st.Args.Test {
		city = "test"
	}

	uri := fmt.Sprintf("/ps/hosts/%s/%s/apps/%d.%s/", city, st.Env.Hostname, dtsApplId, st.Env.DtsInstance)
	updateKeys, err := st.DtsApp.Push(uri)
	if err != nil {
		return
	}

	Log.Println(updateKeys)
	return
}

// Init external git dir and add accessible files
func (st *State) gitInit() error {
	gitDir := joinPaths(st.Env.DtsDir, st.Env.Instance)
	Log.Println("gitInit with env:", st.Env.WorkTree, gitDir)
	b, err := dts.Init(st.Env.WorkTree, gitDir)
	if err != nil {
		return err
	}

	output := rmEscape.Replace(string(b))
	Log.Printf("%s\n", output)

	st.Files = &Files{}
	err = st.Files.walk(st.Env.WorkTree)
	if err != nil {
		return err
	}
	log.Printf("\nAccessible: %+q\nGtSize: %+q\nUnReadable: %+q\n", st.Files.Accessible, st.Files.GtSize, st.Files.UnReadable)
	log.Println("Symlinks:", st.Files.Symlinks)

	// Add & commit
	b, err = dts.AddNCommit(st.Env.WorkTree, gitDir, st.Files.Accessible)
	if err != nil {
		return err
	}

	output = rmEscape.Replace(string(b))
	Log.Println(output)

	return err
}

func (st *State) Status() {
	st.TApp = &etcd.App{}
	ok, err := st.config.FetchAppByInstance(st.Env.Instance, st.TApp)
	st.checkError(err)

	if !ok {
		st.checkError(errExtractingTargetApp)
	}

	// Check dts config with target application
	v, ok := st.DtsApp.DtsSettings.AppList[st.Env.Instance]
	if ok {
		if !v.Enabled {
			// check if target app supporting versioning
			if v.WorkTree != v.AppDir {
				st.Env.WorkTree, err = resolveCurrentVersion(v.AppDir)
				st.checkError(err)

				// if work-tree that symlink points to not matches with one obtained from registry host
				// we consider that a new version of target app was deployed
				if st.Env.WorkTree != v.WorkTree {
					Log.Println("new version of target app was deployed, redeploying...")

					st.Env.AppDir = v.AppDir

					// remove instance from app_list
					delete(st.DtsApp.DtsSettings.AppList, st.Env.Instance)
					// remove measurement from emon_json
					st.DtsApp.EmonJson.RemoveMeasurementByInstance(st.Env.Instance)
					Log.Printf("%s was removed from dts app_list\n", st.Env.Instance)

					// remove git files
					err = removeGitDir(v.GitDir)
					st.checkError(err)

					Log.Printf("instance \"%s\" completely removed\n", st.Env.Instance)

					// init new instance
					st.checkError(st.init())

					st.checkError(st.logJson())
					// early exit
					os.Exit(0)
				}
			} else {
				st.checkError(errInstanceDisabled)
			}
		}

		st.Env.WorkTree = v.WorkTree
		st.Env.AppDir = v.AppDir

		//if st.TApp.AppDir != st.Env.WorkTree {
		//	st.checkError(ErrWorkTreeNotMatch)
		//}

		if st.TApp.ApplicationName != v.AppName {
			st.checkError(errAppNameNotMatch)
		}

		st.MFiles, err = dts.Numstat(st.Env.WorkTree, v.GitDir)
		st.checkError(err)
	} else {
		st.checkError(errInstanceIsNotExist)
	}
}

// Telegraf output status string in telegraf format
func (st *State) Telegraf() {
	output := st.MFiles.Telegraf(st.DtsApp.DtsSettings.AppList[st.Env.Instance].AppName)
	for i := 0; i < len(output); i++ {
		Log.Println(output[i])
		fmt.Println(output[i])
	}
}

// LogJson marshal current state into json format and write it into default json log
func (st *State) LogJson() {
	st.checkError(st.logJson())
}

// Update dts app struct, combine next function (SetDtsSettings, SetEmonJson, SetDtsApp)
func (st *State) setDtsApp() {
	st.DtsApp.DtsSettings.SetDtsSettings(st.Env.AppDir, st.TApp.ApplicationName, st.Env.WorkTree, st.Env.DtsDir, st.Env.Instance)
	st.DtsApp.EmonJson.SetEmonJson(dtsApplId, dtsAppName, st.Env.DtsDir, st.Env.Instance)
	st.DtsApp.SetDtsApp(strconv.Itoa(dtsApplId), dtsAppName, st.TApp.Stand, st.DtsApp.DtsSettings, st.DtsApp.EmonJson)
}

// Get etcd url based on short host name
func getEtcdUrl(sName string) string {
	var url string
	testZone := regexp.MustCompile(`^([a-z]{2,4}-?){3}\d+[a-z]$`)

	if testZone.Match([]byte(sName)) {
		url = fmt.Sprintf("http://%s%s:%s", etcdTestUrlPref, sName[len(sName)-1:], etcdDefaultPort)
	} else {
		url = fmt.Sprintf("http://%s:%s", etcdGreenUrl, etcdDefaultPort)
	}

	return url
}

// getInstance return crc32 hash of a given path, is equivalent to shell command: echo "/path/to/app" | cksum
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
