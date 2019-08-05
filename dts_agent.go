package main

import (
	"./crc"
	"./dts"
	"./etcd"
	"errors"
	"fmt"
	"go.etcd.io/etcd/client"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	dtsId = "5118"
)

var (
	ErrInstanceNotExist = errors.New("instance do not exists in etcd")
	ErrInstanceExist    = errors.New("instance already exists")
	ErrAppDirNotMatch   = errors.New("app dirs do not match")
)

func main() {
	// Test arguments
	args := []string{
		"--action", "status",
		//"--dts-dir", "C:/gitdir",
		//"--work-tree", "C:/golang-webapp",
		"--instance", "3082379100",
		"-t",
		//"--help",
		//"-V",
	}

	// Get flag options
	//opts, err := GetParser(args)
	opts, err := GetParser(args)
	checkError(err)

	// Check passed arguments
	err = opts.CheckArgs()
	checkError(err)

	// Generate part of dts settings based on passed arguments
	dtsDir, etcdUrl, sHostName, err := opts.ObtainDtsVars()
	checkError(err)

	var instance, dtsInstance string
	dtsInstance = getInstance(dtsDir)

	var config etcd.Etcd
	var dtsApp, tApp etcd.App

	err = etcd.FetchApps(&config, etcdUrl, sHostName)
	checkError(err)

	if opts.Action == "init" {
		versionLess := leaveCurrent(opts.WorkTree)
		instance = getInstance(versionLess)
		log.Println(versionLess, instance)

		ok, err := config.ExtractAppByInstance(instance, &tApp)
		checkError(err)

		if !ok {
			checkError(ErrInstanceNotExist)
		}

		if tApp.AppDir != versionLess {
			checkError(ErrAppDirNotMatch)
		}

		log.Println("dts instance:", dtsInstance)
		ok, err = config.ExtractAppByInstance(dtsInstance, &dtsApp)
		checkError(err)

		if ok {
			if _, ok = dtsApp.DtsSettings.AppList[instance]; ok {
				checkError(ErrInstanceExist)
			}
			log.Printf("initialize with: %s, %s, %s, %s", tApp.ApplicationName, opts.WorkTree, dtsDir, instance)
			dtsApp.DtsSettings.SetDtsSettings(tApp.ApplicationName, opts.WorkTree, dtsDir, instance)
			dtsApp.EmonJson.SetEmonJson(dtsId, opts.DtsDir, instance)
			log.Printf("added app to applist and emon json: %+v\n", dtsApp)
		} else {
			dtsSettings := &etcd.DtsSettings{}
			emonJson := &etcd.EmonJson{}
			dtsSettings.SetDtsSettings(tApp.ApplicationName, opts.WorkTree, opts.DtsDir, instance)
			emonJson.SetEmonJson(dtsId, opts.DtsDir, instance)
			dtsApp.SetDtsApp(dtsId, "DTS", "GF", dtsSettings, emonJson)
			log.Printf("generated new app: %+v\n", dtsApp)
		}

		err = taskInit(opts.WorkTree, dtsDir, instance)
		checkError(err)

		cfg := client.Config{
			Endpoints:               []string{etcdUrl},
			Transport:               client.DefaultTransport,
			HeaderTimeoutPerRequest: time.Second,
		}

		c, err := client.New(cfg)
		checkError(err)

		kApi := client.NewKeysAPI(c)

		uri := fmt.Sprintf("/ps/hosts/%s/apps/%s.%s/", sHostName, dtsId, dtsInstance)
		err = dtsApp.WriteApp(uri, kApi)
		checkError(err)

	} else if opts.Action == "status" {
		instance = opts.Instance
		ok, err := config.ExtractAppByInstance(instance, &tApp)
		checkError(err)

		if !ok {
			checkError(ErrInstanceNotExist)
		}

		ok, err = config.ExtractAppByInstance(dtsInstance, &dtsApp)
		checkError(err)

		if !ok {
			checkError(ErrInstanceNotExist)
		}

		// Check dts config with target application
		if v, ok := dtsApp.DtsSettings.AppList[instance]; ok && tApp.AppDir == leaveCurrent(v.WorkTree) && tApp.ApplicationName == v.AppName {
			repo, err := dts.Open(v.WorkTree, v.GitDir)
			checkError(err)

			diffs, err := dts.Diff(repo)
			checkError(err)

			// Telegraf output
			fmt.Println(diffs.Telegraf(v.AppName))
		} else {
			checkError(ErrInstanceNotExist)
		}
	}
}

//func tInit() {
//
//}

func taskInit(workTree, dtsDir, instance string) error {
	wt, err := dts.Init(workTree, dtsDir, instance)
	checkError(err)

	accessible, unReadable, gtSize, symlinks, err := Walk(workTree)
	checkError(err)
	log.Printf("accessible: %v, unReadable: %v, gtSize: %v, symlinks: %v\n", accessible, unReadable, gtSize, symlinks)
	// Add & commit
	return dts.Commit(workTree, accessible, wt)

}

func leaveCurrent(workTree string) string {
	path := strings.Split(workTree, "/")
	if path[len(path)-1] == "current" {
		return strings.Join(path[:len(path)-1], "/")
	}
	return strings.Join(path, "/")
}

func getInstance(workTree string) string {
	return strconv.FormatUint(uint64(crc.CkSum(workTree+"\n")), 10)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
