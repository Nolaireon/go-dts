package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	version         = "0.2"
	dtsDefaultDir   = "/data/usnmp/go-dts"
	dtsCustomDir    = "C:/gitdir"
	etcdDefaultPort = "2500"
	etcdCustomPort  = "2379"
	etcdTestUrl     = "vlg-mon-app1"
	etcGreenUrl     = "influx.megafon.ru"
	etcdCustomUrl   = "centos.emink.net"
)

var (
	opts   = &Options{}
	parser = flags.NewParser(opts, flags.Default)
)

type Options struct {
	Help     helpOptions `group:"Help Options"`
	Action   string      `short:"a" long:"action" description:"init or status" choice:"init" choice:"status" required:"true"`
	WorkTree string      `short:"w" long:"work-tree" description:"path to application"`
	DtsDir   string      `short:"d" long:"dts-dir" description:"path to dts directory"`
	Instance string      `short:"i" long:"instance" description:"crc of application path"`
	Test     bool        `short:"t" long:"test" description:"use test args"`
}

type helpOptions struct {
	Help    bool `short:"h" long:"help" description:"show help"`
	Version bool `short:"V" long:"version" description:"show version"`
}

func init() {
	parser.Usage = "--action=[init,status] [--work-tree [--dts-dir], --instance]"
}

func GetParser(args []string) (*Options, error) {
	_, err := parser.ParseArgs(args)
	return opts, err
}

func (opts *Options) CheckArgs() (err error) {
	if opts.Help.Help {
		parser.WriteHelp(os.Stderr)
		os.Exit(0)
	}

	if opts.Help.Version {
		log.Printf("%s version: %s", os.Args[0], version)
		os.Exit(0)
	}

	switch opts.Action {
	case "init":
		if len(opts.WorkTree) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "work-tree required for init action"}
		}
	case "status":
		if len(opts.Instance) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "instance required for status action"}
		}
	default:
		err = &flags.Error{Type: flags.ErrInvalidChoice, Message: "invalid action value, should be [init, status]"}
	}

	return
}

func (opts *Options) ObtainDtsVars() (dtsDir, etcdUrl, sName string, err error) {
	if opts.Test {
		sName = "vlg-lbrt-app1d"
		etcdUrl = fmt.Sprintf("http://%s:%s", etcdCustomUrl, etcdCustomPort)
		dtsDir = dtsCustomDir
	} else {
		sName, err = getShortHostName()
		if err != nil {
			return
		}

		etcdUrl = getEtcdUrl(sName)

		if len(opts.DtsDir) == 0 {
			dtsDir = dtsDefaultDir
		} else {
			dtsDir = opts.DtsDir
		}
	}
	return
}

// GetShortHostName return short name of domain
func getShortHostName() (sName string, err error) {
	hostName, err := os.Hostname()
	if err != nil {
		return
	}

	sName = strings.ReplaceAll(hostName, ".megafon.ru", "")
	return
}

// Get etcd url based on host short name
func getEtcdUrl(sName string) string {
	var url string
	zoneTest := regexp.MustCompile(`^([a-z]{2,4}-?){3}\d+[a-z]$`)

	if zoneTest.Match([]byte(sName)) {
		url = fmt.Sprintf("http://%s%s:%s", etcdTestUrl, sName[len(sName)-1:], etcdDefaultPort)
	} else {
		url = fmt.Sprintf("http://%s:%s", etcGreenUrl, etcdDefaultPort)
	}

	return url
}
