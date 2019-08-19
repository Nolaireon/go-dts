package main

import (
	"./task"
)

var (
	ex = []string{"7001"}
)

func main() {
	// Test arguments
	args := []string{
		"--action", "deploy",
		//"--dts-dir", "C:/gitdir",
		//"--work-tree", "C:/test",
		//"--instance", "1391523444",
		//"--instance", "3082379100",
		"-t",
		//"--help",
		//"-V",
	}

	st := &task.State{}
	st.NewParser()
	st.ParseArgs(args)

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
