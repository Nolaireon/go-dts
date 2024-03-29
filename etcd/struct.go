package etcd

// ETCD structure
type Etcd struct {
	Action string `json:"action"`
	Node   Node   `json:"node"`
}

type Node struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Nodes []Node `json:"nodes"`
}

// Application structure
type App struct {
	AppDir          string       `json:"app_dir,omitempty"`
	ApplId          string       `json:"appl_id,omitempty"`
	ApplicationName string       `json:"application_name,omitempty"`
	CurrentVersion  string       `json:"current_version,omitempty"`
	Instance        string       `json:"instance,omitempty"`
	PidFile         string       `json:"pid_file,omitempty"`
	ProductName     string       `json:"product_name,omitempty"`
	Stand           string       `json:"stand,omitempty"`
	DtsSettings     *DtsSettings `json:"dts_settings,omitempty"`
	EmonJson        *EmonJson    `json:"emon_json,omitempty"`
}

// DTS settings structure
type DtsSettings struct {
	AppList map[string]*Instance `json:"app_list"`
	Updated string               `json:"updated"`
}

type Instance struct {
	AppDir   string `json:"app_dir"`
	AppName  string `json:"app_name"`
	Enabled  bool   `json:"enabled"`
	GitDir   string `json:"git_dir"`
	WorkTree string `json:"work_tree"`
	//LockFile string `json:"lock_file"`
}

// emon_json structure for DTS
type EmonJson struct {
	ApplId       int            `json:"appl_id"`
	Description  string         `json:"description"`
	Measurements []*Measurement `json:"measurements"`
	Product      string         `json:"product"`
	Service      string         `json:"service"`
}

type Measurement struct {
	Name          string         `json:"name"`
	Configuration *Configuration `json:"configuration"`
}

type Configuration struct {
	Commands   []string `json:"commands"`
	DataFormat string   `json:"data_format"`
	Interval   string   `json:"interval"`
	Timeout    string   `json:"timeout"`
	Type       string   `json:"type"`
}
