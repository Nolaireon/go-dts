package task

import (
	"../dts"
	"../etcd"
)

// Command-line arguments
type Arguments struct {
	Help     helpOptions `group:"Help Options" json:"-"`
	Action   string      `short:"a" long:"action" description:"init, status or deploy" choice:"init" choice:"status" choice:"deploy" required:"true" json:"action,omitempty"`
	WorkTree string      `short:"w" long:"work-tree" description:"path to application" json:"work_tree,omitempty"`
	//DtsDir   string      `short:"d" long:"dts-dir" description:"path to dts directory" json:"dts_dir,omitempty"`
	Instance string `short:"i" long:"instance" description:"crc of application path" json:"instance,omitempty"`
	Test     bool   `short:"t" long:"test" description:"use test args" json:"test,omitempty"`
}

type helpOptions struct {
	Help    bool `short:"h" long:"help" description:"show help"`
	Version bool `short:"V" long:"version" description:"show version"`
}

// Contain current state
type State struct {
	config *etcd.Etcd
	DtsApp *etcd.App    `json:"dts_app"`
	TApp   *etcd.App    `json:"t_app"`
	Files  *Files       `json:"files,omitempty"`
	MFiles *dts.MFiles  `json:"m_files,omitempty"`
	Args   *Arguments   `json:"args"`
	Env    *Environment `json:"env"`
	Time   string       `json:"time"`
	Err    string       `json:"error,omitempty"`
}

type Files struct {
	Accessible []string    `json:"accessible,omitempty"`
	UnReadable []string    `json:"unreadable,omitempty"`
	GtSize     []string    `json:"gt_size,omitempty"`
	Symlinks   [][2]string `json:"symlinks,omitempty"`
}

// Environment contain variables generated in runtime before main task execution
type Environment struct {
	WorkTree    string `json:"work_tree,omitempty" yaml:"work_tree,omitempty"`
	AppDir      string `json:"app_dir,omitempty" yaml:"app_dir,omitempty"`
	Instance    string `json:"instance,omitempty" yaml:"instance,omitempty"`
	EtcdUrl     string `json:"etcd_url,omitempty" yaml:"etcd_url,omitempty"`
	DtsDir      string `json:"dts_dir,omitempty" yaml:"dts_dir,omitempty"`
	DtsInstance string `json:"dts_instance,omitempty" yaml:"dts_instance,omitempty"`
	Hostname    string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
}

type ExcludedApps struct {
	ExcludedApps []string `yaml:"excluded_apps"`
}

//type Config struct {
//	ExcludedApps []string     `yaml:"excluded_apps"`
//	Args         *Arguments   `yaml:"args,omitempty"`
//	Env          *Environment `yaml:"env,omitempty"`
//}
