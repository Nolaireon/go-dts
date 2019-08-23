package main

import (
	"./task"
)

var (
	ex = []string{
		// InfluxDB, Kapacitor, Telegraf, Grafana
		"22011001", "22011002", "22011003", "22011004",
		// uWSGI, MySql, Relay-server, Etcd, Nginx Unit, Confd, MonDB, Redis, Showcase
		"99000049", "99000050", "99000051", "99000052", "99000053", "99000055", "99000056", "99000057", "99000058",
		// Prometheus, Alertmanager, Alertstrap, Alertsender
		"99000010", "99000011", "99000200", "99000201",
		// openfs processes (IUM), collectors IUM, WEBSSO_APP, WEBSSO_ARM, Consul, PIM, Microservices EAPI, GitLab
		"99000060", "99000061", "99000070", "99000071", "99000072", "99000073", "99000074", "99000080",
		// Kafka for ELOG, ClickHouse, SSO_APP, Docker, Cubernetes, Kraken, Stale dispatcher, FMC_EAPI
		"99000110", "99000120", "99000130", "99000150", "99000160", "99000200", "99000220", "99000400",
		// usnmp int, ELK...
		"9026", "09035001", "09035002", "09035003", "09035004", "09035005", "09035006", "09035008", "09035009", "09035010",
	}
)

func main() {
	// Test arguments
	//args := []string{
	//	"--action", "deploy",
	//	//"--dts-dir", "C:/gitdir",
	//	//"--work-tree", "C:/test",
	//	//"--instance", "1391523444",
	//	//"--instance", "3082379100",
	//	"-t",
	//	//"--help",
	//	//"-V",
	//}

	st := &task.State{}
	st.NewParser()
	st.ParseArgs()

	st.Fetch()

	switch st.Opts.Action {
	case "init":
		st.Init()
		st.LogJson()
	case "status":
		st.Status()
		st.Telegraf()
		st.LogJson()
	case "deploy":
		st.Deploy(ex)
	}
}
