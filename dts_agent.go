package main

import (
	"./modules"
	"fmt"
	"github.com/jessevdk/go-flags"
	"go.etcd.io/etcd/client"
	"io"
	"log"
	"os"
	"time"
)

const (
	Version  = "0.1"
	ApplId   = "5118"
	ApplHash = "1435232203"
)

var options struct {
	Help     HelpOptions `group:"Help Options"`
	Action   string      `short:"a" long:"action" description:"init or status" choice:"init" choice:"status" required:"true"`
	WorkTree string      `short:"w" long:"work-tree" description:"path to application"`
	GitDir   string      `short:"g" long:"git-dir" description:"path to git directory"`
	Instance uint32      `short:"i" long:"instance" description:"crc of application path"`
}

type HelpOptions struct {
	Help    bool `short:"h" long:"help" description:"show help"`
	Version bool `short:"V" long:"version" description:"show version"`
}

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

var (
	opts     = options
	parser   = flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash)
	logger   = log.New(&writer{os.Stdout, time.RFC3339 + " "}, "", 0)
	instance = "582108783"
)

func main() {
	args := []string{
		"--action", "status",
		"--git-dir", "C:\\gitdir",
		"--work-tree", "C:\\test",
		"--instance", "582108783",
		//"--help",
		//"-V",
	}

	parser.Usage = "--action=[init,status] [--work-tree [--git-dir], --instance]"

	_, err := parser.ParseArgs(args)
	checkError(err)

	if opts.Help.Help {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if opts.Help.Version {
		fmt.Printf("%s version %s\n", os.Args[0], Version)
		os.Exit(0)
	}

	if opts.Action == "init" {
		if len(opts.WorkTree) == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "work-tree required for init action."}
			checkError(err)
		}

		w, err := modules.Init(opts.WorkTree, opts.GitDir, instance)
		checkError(err)

		accessible, _, _, _, err := modules.Walk(opts.WorkTree)
		//Add & commit
		err = modules.Commit(opts.WorkTree, accessible, w)
		checkError(err)
	} else if opts.Action == "status" {
		if opts.Instance == 0 {
			err = &flags.Error{Type: flags.ErrCommandRequired, Message: "instance required for status action"}
			checkError(err)
		}

		repo, err := modules.Open(opts.WorkTree, opts.GitDir, instance)
		checkError(err)
		diffs, err := modules.Diff(repo)
		checkError(err)
		if diffs != nil {
			fmt.Println("not nil:", diffs)
		} else {
			fmt.Println("diffs nil")
		}
	}
	os.Exit(0)
	data := []byte(`{"action":"get","node":{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/application_name","value":"TELEGRAF","modifiedIndex":74192,"createdIndex":74192},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/current_version","value":"1.11.0","modifiedIndex":74187,"createdIndex":74187},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/info_json","value":"{\n    \"product\": {\n        \"group\": \"MON\",\n        \"name\": \"\",\n        \"Version\": \"1.11.0\"\n    },\n    \"application\": {\n        \"name\": \"TELEGRAF\",\n        \"appl_id\": \"22011003\",\n        \"Version\": \"1.11.0\",\n        \"date\": \"\",\n        \"build\": {\n            \"number\": \"28\"\n        }\n    }\n}\n","modifiedIndex":74209,"createdIndex":74209},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/stand","value":"GF","modifiedIndex":74202,"createdIndex":74202},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/appl_id","value":"22011003","modifiedIndex":74180,"createdIndex":74180},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/instance","value":"1","modifiedIndex":74184,"createdIndex":74184},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/pid_file","value":"/data/usnmp/telegraf-app/logs/telegraf.pid","modifiedIndex":7233,"createdIndex":7233},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/product_name","value":"","modifiedIndex":74197,"createdIndex":74197},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/22011003.1202610149/app_dir","value":"/data/usnmp/telegraf-app","modifiedIndex":74175,"createdIndex":74175}],"modifiedIndex":7201,"createdIndex":7201},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/5118.1435232203","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/5118.1435232203/appl_id","value":"5118","modifiedIndex":82329,"createdIndex":82329},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/5118.1435232203/emon_json","value":"{\n    \"appl_id\": \"5118\", \n    \"description\": \"DTS\", \n    \"measurements\": [\n        {\n            \"configuration\": {\n                \"commands\": [\n                    \"/usr/bin/python /data/usnmp/dts-app/dts_agent.py status --instance 582108783\"\n                ], \n                \"data_format\": \"influx\", \n                \"interval\": \"5m\", \n                \"timeout\": \"30s\", \n                \"type\": \"exec\"\n            }, \n            \"name\": \"data-tracking-system\"\n        }\n    ], \n    \"product\": \"DTS\"\n}","modifiedIndex":82330,"createdIndex":82330},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/5118.1435232203/application_name","value":"DTS","modifiedIndex":82326,"createdIndex":82326},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/5118.1435232203/dts_settings","value":"{\n    \"app_list\": {\n        \"582108783\": {\n            \"app_name\": \"BRT_SRV\", \n            \"enabled\": \"True\", \n            \"git_dir\": \"/data/usnmp/dts-app/582108783/\", \n            \"lock_file\": \"/data/usnmp/dts-app/582108783/.lock\", \n            \"work_tree\": \"/data/brt/BRT/current/\"\n        }\n    }, \n    \"updated\": \"2019-07-18 11:27:58.625935\"\n}","modifiedIndex":82327,"createdIndex":82327},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/5118.1435232203/stand","value":"GF","modifiedIndex":82328,"createdIndex":82328}],"modifiedIndex":82326,"createdIndex":82326},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/info_json","value":"{\n    \"application\": {\n        \"appl_id\": \"8098\", \n        \"name\": \"NTF_SRV\", \n        \"Version\": \"1.5.1\"\n    }, \n    \"product\": {\n        \"group\": \"NS\", \n        \"name\": \"NWM_OCS\", \n        \"Version\": \"2.4.2.10\"\n    }\n}","modifiedIndex":77981,"createdIndex":77981},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/instance","value":"ntf_srv.11","modifiedIndex":82336,"createdIndex":82336},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/product_name","value":"NWM_OCS","modifiedIndex":41881,"createdIndex":41881},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/stand","value":"GF","modifiedIndex":22194,"createdIndex":22194},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/app_dir","value":"/data/brt/NTF_SRV","modifiedIndex":6419,"createdIndex":6419},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/appl_id","value":"8098","modifiedIndex":77986,"createdIndex":77986},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/application_name","value":"NTF_SRV","modifiedIndex":6425,"createdIndex":6425},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/current_version","value":"2.4.2.10","modifiedIndex":77984,"createdIndex":77984},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/pid_file","value":"/data/brt/NTF_SRV/current/bin/ntf_srv.1.pid","modifiedIndex":43661,"createdIndex":43661},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/8098.1607630698/history","value":"[\n    {\n        \"date\": \"2019-03-04 07:31:46\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"asapozhn\", \n        \"Version\": \"2.3.2.3\"\n    }, \n    {\n        \"date\": \"2019-04-01 07:08:10\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.9\"\n    }, \n    {\n        \"date\": \"2019-04-04 08:08:06\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"asapozhn\", \n        \"Version\": \"2.3.2.10\"\n    }, \n    {\n        \"date\": \"2019-04-08 08:05:45\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.11\"\n    }, \n    {\n        \"date\": \"2019-04-24 06:09:31\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.14\"\n    }, \n    {\n        \"date\": \"2019-06-24 12:32:10\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.6\"\n    }, \n    {\n        \"date\": \"2019-06-24 13:05:00\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.6\"\n    }, \n    {\n        \"date\": \"2019-06-24 13:07:09\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.15\"\n    }, \n    {\n        \"date\": \"2019-07-04 06:53:15\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.8\"\n    }, \n    {\n        \"date\": \"2019-07-10 06:42:40\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.10\"\n    }\n]","modifiedIndex":77983,"createdIndex":77983}],"modifiedIndex":6419,"createdIndex":6419},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/application_name","value":"SNMP_INT_SERVER","modifiedIndex":39559,"createdIndex":39559},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/info_json","value":"{\n  \"product\":{\n    \"group\":\"MON\",\n    \"name\":\"EMON\",\n    \"Version\":\"1.3.1\"\n  },\n  \"application\":{\n    \"name\":\"SNMP_INT_SERVER\",\n    \"appl_id\":\"9026\",\n    \"Version\":\"2.0.7\",\n    \"date\":\"2018-05-25T08:20:24Z\",\n    \"build\":{\n      \"number\":\"4\"\n    }\n  }\n}","modifiedIndex":39568,"createdIndex":39568},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/appl_id","value":"9026","modifiedIndex":39554,"createdIndex":39554},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/current_version","value":"2.0.7","modifiedIndex":39556,"createdIndex":39556},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/instance","value":"","modifiedIndex":9207,"createdIndex":9207},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/product_name","value":"EMON","modifiedIndex":39562,"createdIndex":39562},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/stand","value":"GF","modifiedIndex":39565,"createdIndex":39565},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/9026.2373578764/app_dir","value":"/data/usnmp/snmp_int-app","modifiedIndex":39551,"createdIndex":39551}],"modifiedIndex":2804,"createdIndex":2804},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/instance","value":"1","modifiedIndex":79488,"createdIndex":79488},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/stand","value":"GF","modifiedIndex":79509,"createdIndex":79509},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/app_dir","value":"/data/usnmp/confd-app","modifiedIndex":79478,"createdIndex":79478},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/info_json","value":"{\n    \"product\": {\n        \"group\": \"MON\",\n        \"name\": \"\",\n        \"Version\": \"0.18.0\"\n    },\n    \"application\": {\n        \"name\": \"CONFD\",\n        \"appl_id\": \"99000055\",\n        \"Version\": \"0.18.0\",\n        \"date\": \"\",\n        \"build\": {\n            \"number\": \"48\"\n        }\n    }\n}\n","modifiedIndex":79514,"createdIndex":79514},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/current_version","value":"0.18.0","modifiedIndex":79493,"createdIndex":79493},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/pid_file","value":"/data/usnmp/confd-app/logs/confd.pid","modifiedIndex":80910,"createdIndex":80910},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/product_name","value":"","modifiedIndex":79504,"createdIndex":79504},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/appl_id","value":"99000055","modifiedIndex":79483,"createdIndex":79483},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/99000055.677870805/application_name","value":"CONFD","modifiedIndex":79498,"createdIndex":79498}],"modifiedIndex":10629,"createdIndex":10629},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/pid_file","value":"/data/brt/BRT/current/bin/brt_srv.M0.pid","modifiedIndex":43660,"createdIndex":43660},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/app_dir","value":"/data/brt/BRT","modifiedIndex":41804,"createdIndex":41804},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/appl_id","value":"7001","modifiedIndex":77286,"createdIndex":77286},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/product_name","value":"NWM_OCS","modifiedIndex":41810,"createdIndex":41810},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/application_name","value":"BRT_SRV","modifiedIndex":41812,"createdIndex":41812},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/current_version","value":"2.4.2.10","modifiedIndex":77287,"createdIndex":77287},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/history","value":"[\n    {\n        \"date\": \"2019-03-22 11:23:47\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"vozornin\", \n        \"Version\": \"2.3.2.8\"\n    }, \n    {\n        \"date\": \"2019-04-01 07:05:55\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.9\"\n    }, \n    {\n        \"date\": \"2019-04-04 07:22:34\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"asapozhn\", \n        \"Version\": \"2.3.2.10\"\n    }, \n    {\n        \"date\": \"2019-04-08 07:59:57\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.11\"\n    }, \n    {\n        \"date\": \"2019-04-24 06:06:38\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.3.2.14\"\n    }, \n    {\n        \"date\": \"2019-06-24 12:27:24\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.6\"\n    }, \n    {\n        \"date\": \"2019-06-24 13:04:58\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.6\"\n    }, \n    {\n        \"date\": \"2019-07-04 06:53:15\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.8\"\n    }, \n    {\n        \"date\": \"2019-07-04 20:17:16\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.9\"\n    }, \n    {\n        \"date\": \"2019-07-04 23:49:17\", \n        \"hostname\": \"vlg-bss-ansib1d.megafon.ru\", \n        \"user\": \"mvlasenk\", \n        \"Version\": \"2.4.2.10\"\n    }\n]","modifiedIndex":77290,"createdIndex":77290},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/stand","value":"GF","modifiedIndex":41795,"createdIndex":41795},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/instance","value":"M0","modifiedIndex":41801,"createdIndex":41801},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/7001.582108783/info_json","value":"{\n    \"application\": {\n        \"appl_id\": \"7001\", \n        \"name\": \"BRT_SRV\", \n        \"Version\": \"2.4.4\"\n    }, \n    \"product\": {\n        \"group\": \"NS\", \n        \"name\": \"NWM_OCS\", \n        \"Version\": \"2.4.2.10\"\n    }\n}","modifiedIndex":77285,"createdIndex":77285}],"modifiedIndex":41795,"createdIndex":41795},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429","dir":true,"nodes":[{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/appl_id","value":"09035006","modifiedIndex":2788,"createdIndex":2788},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/application_name","value":"LOGS_AGENT","modifiedIndex":2784,"createdIndex":2784},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/current_version","value":"7.7.0","modifiedIndex":2789,"createdIndex":2789},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/info_json","value":"{\"application\": {\"date\": \"2018-05-18T10:07:39Z\", \"Version\": \"6.2.3\", \"name\": \"LOGS_AGENT\", \"appl_id\": \"09035006\", \"build\": {\"number\": \"1124\"}}, \"product\": {\"Version\": \"7.7.0\", \"group\": \"BIN\", \"name\": \"ELOG\"}}","modifiedIndex":2794,"createdIndex":2794},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/instance","value":"1","modifiedIndex":2787,"createdIndex":2787},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/stand","value":"GF","modifiedIndex":22200,"createdIndex":22200},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/app_dir","value":"/data/elog/filebeat","modifiedIndex":2795,"createdIndex":2795},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/history","value":"[{\"date\": \"2018-07-05 12:06:52\", \"Version\": \"7.7.0\", \"user\": \"erubtsov\", \"hostname\": \"vlg-bss-ansib1d.megafon.ru\"}]","modifiedIndex":2790,"createdIndex":2790},{"key":"/ps/hosts/vlg/vlg-lbrt-app1d/apps/09035006.470223429/product_name","value":"ELOG","modifiedIndex":2791,"createdIndex":2791}],"modifiedIndex":2784,"createdIndex":2784}],"modifiedIndex":2784,"createdIndex":2784}}`)

	var config modules.Etcd
	var dtsApp modules.App
	//var tApp modules.App

	err = modules.Fetch(data, &config)
	checkError(err)

	err = config.ExtractApp("1435232203", &dtsApp)
	checkError(err)

	//err = config.ExtractApp("582108783", &tApp)
	//checkError(err)

	cfg := client.Config{
		Endpoints:               []string{"http://centos.emink.net:2379"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	checkError(err)

	kapi := client.NewKeysAPI(c)
	//uri := "/ps/hosts/vlg-lbrt-app1d/apps/7001.582108783/"
	uri := "/ps/hosts/vlg-lbrt-app1d/apps/5118.1435232203/"
	//err = tApp.WriteApp(uri, kapi)
	err = dtsApp.WriteApp(uri, kapi)
	checkError(err)

	fmt.Println(modules.CkSum("/data/hrs/HRS_RT\n"))

	accessible, unReadable, gtSize, symlinks, err := modules.Walk("C:/gradle5.4.1")
	checkError(err)

	fmt.Println("accessible:", len(accessible))
	fmt.Println("unReadable:", unReadable)
	fmt.Println("gtSize:", len(gtSize))
	fmt.Println("symlinks:", symlinks)
}

func checkError(err error) {
	if err != nil {
		logger.Fatal(err)
	}
}
