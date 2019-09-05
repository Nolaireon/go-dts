package main

import (
	"./task"
)

func main() {
	// empty args mean that ParseArgs will parse arguments from os.Args
	args := make([]string, 0)
	// Test arguments
	//args := []string{
	//	"--action", "init",
	//	//"--dts-dir", "C:/gitdir",
	//	"--work-tree", "C:\\test",
	//	//"--instance", "3701349221",
	//	//"--instance", "3082379100",
	//	"-t",
	//	//"--help",
	//	//"-V",
	//}

	st := &task.State{}
	st.ParseArgs(args)

	st.PrepareEnv()

	st.Fetch()

	switch st.Args.Action {
	case "init":
		st.PlainInit()
		st.LogJson()
	case "status":
		st.Status()
		st.Telegraf()
		st.LogJson()
	case "deploy":
		st.Deploy()
	}
}
