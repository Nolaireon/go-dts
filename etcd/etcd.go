package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"go.etcd.io/etcd/client"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var kApi client.KeysAPI

func SetEtcdApi(url string) (err error) {
	cfg := client.Config{
		Endpoints:               []string{url},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	var c client.Client
	c, err = client.New(cfg)
	if err != nil {
		return
	}

	kApi = client.NewKeysAPI(c)
	return
}

func (config *Etcd) FetchConfig(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &config)
	return
}

// CollectApps returns a list of apps obtained from the registry host, excluding apps from config exclude_apps.yml
func (config *Etcd) CollectApps(excludedApps []string) (apps [][]string) {
	for _, value := range config.Node.Nodes {
		key := strings.Split(value.Key, "/")
		idDotHash := strings.Split(key[len(key)-1], ".")

		doAppend := true
		for i := 0; i < len(excludedApps); i++ {
			if idDotHash[0] == excludedApps[i] {
				doAppend = false
				break
			}
		}

		if doAppend {
			apps = append(apps, idDotHash)
		}
	}

	return
}

// ok == true means that app was found by given instance
func (config *Etcd) FetchAppByInstance(instance string, app *App) (ok bool, err error) {
	replacer := strings.NewReplacer("\n", "", "    ", "")

	for _, value := range config.Node.Nodes {
		key := strings.Split(value.Key, "/")
		idDotHash := strings.Split(key[len(key)-1], ".")
		if idDotHash[1] == instance {
			for _, v := range value.Nodes {
				switch k := strings.Split(v.Key, "/"); k[len(k)-1] {
				case "dts_settings":
					tempV := replacer.Replace(v.Value)
					if err = json.Unmarshal([]byte(tempV), &app.DtsSettings); err != nil {
						return
					}
				case "emon_json":
					tempV := replacer.Replace(v.Value)
					if err = json.Unmarshal([]byte(tempV), &app.EmonJson); err != nil {
						return
					}
				//case "app_dir":
				//	app.AppDir = v.Value
				//case "appl_id":
				//	app.ApplId = v.Value
				//case "application_name":
				//	app.ApplicationName = v.Value
				//case "current_version":
				//	app.CurrentVersion = v.Value
				//case "history":
				//	app.History = v.Value
				//case "info_json":
				//	app.InfoJson = v.Value
				//case "instance":
				//	app.Instance = v.Value
				//case "pid_file":
				//	app.PidFile = v.Value
				//case "product_name":
				//	app.ProductName = v.Value
				//case "stand":
				//	app.Stand = v.Value
				default:
					refVal := reflect.ValueOf(app).Elem()
					fieldName := getStructFieldKey(k[len(k)-1])
					refVal.FieldByName(fieldName).SetString(v.Value)
				}
			}
			ok = true
		}
	}

	return
}

// Convert key_field to KeyField
func getStructFieldKey(key string) string {
	sl := strings.Split(key, "_")
	for i := 0; i < len(sl); i++ {
		sl[i] = strings.Title(sl[i])
	}

	return strings.Join(sl, "")
}

func (app *App) Push(uri string) (updateKeys []string, err error) {
	v := reflect.ValueOf(app).Elem()
	resp := &client.Response{}
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Tag.Get("json")
		// remove omitempty if exists
		key = strings.Split(key, ",")[0]
		//key = strings.ReplaceAll(key, ",omitempty", "")
		switch v.Field(i).Interface().(type) {
		case string:
			if v.Field(i).Len() > 0 {
				resp, err = kApi.Set(context.Background(), uri+key, v.Field(i).String(), nil)
				if err != nil {
					return
				}

				updateKeys = append(updateKeys, fmt.Sprintf("key %s: %s=%s\n", resp.Action, uri+key, v.Field(i).String()))
			}
		case *EmonJson:
			if cmp.Equal(&app.EmonJson, &EmonJson{}) {
				continue
			}

			var buf []byte
			buf, err = json.MarshalIndent(app.EmonJson, "", "    ")
			if err != nil {
				return
			}

			resp, err = kApi.Set(context.Background(), uri+key, string(buf), nil)
			if err != nil {
				return
			}

			updateKeys = append(updateKeys, fmt.Sprintf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Elem().Interface()))
		case *DtsSettings:
			if cmp.Equal(&app.DtsSettings, &DtsSettings{}) {
				continue
			}

			var buf []byte
			buf, err = json.MarshalIndent(app.DtsSettings, "", "    ")
			if err != nil {
				return
			}

			resp, err = kApi.Set(context.Background(), uri+key, string(buf), nil)
			if err != nil {
				return
			}

			updateKeys = append(updateKeys, fmt.Sprintf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Elem().Interface()))
		}
	}

	return
}

// SetDtsSettings set or update dts_settings struct
func (ds *DtsSettings) SetDtsSettings(appDir, appName, workTree, dtsDir, instance string) {
	gitDir := filepath.Join(dtsDir, instance)
	lockFile := filepath.Join(gitDir, ".lock")
	if ds.AppList == nil {
		ds.AppList = map[string]*Instance{}
	}

	ds.AppList[instance] = &Instance{
		AppDir:   appDir,
		AppName:  appName,
		WorkTree: workTree,
		GitDir:   gitDir,
		Enabled:  true,
		LockFile: lockFile,
	}
	ds.Updated = time.Now().Format(time.RFC3339)
}

// SetEmonJson set or update emon_json struct
func (ej *EmonJson) SetEmonJson(dtsId, dtsAppName, gitDir, instance string) {
	dtsApp := filepath.Join(gitDir, strings.ToLower(dtsAppName))
	command := fmt.Sprintf("%s -a status -i %s", dtsApp, instance)
	ej.ApplId = dtsId
	ej.Description = "data-tracking-system powered on go"
	ej.Measurements = append(ej.Measurements, Measurement{
		Name: "data-tracking-system",
		Configuration: Configuration{
			Commands:   []string{command},
			DataFormat: "influx",
			Interval:   "5m",
			Timeout:    "30s",
			Type:       "exec",
		},
	})
	ej.Product = dtsAppName
	ej.Service = dtsAppName
}

// SetDtsApp set or update dts app struct
func (app *App) SetDtsApp(dtsId, appName, stand string, ds *DtsSettings, ej *EmonJson) {
	app.ApplId = dtsId
	app.ApplicationName = appName
	app.DtsSettings = ds
	app.EmonJson = ej
	app.Stand = stand
}

func (ej *EmonJson) RemoveMeasurementByInstance(instance string) {
	length := len(ej.Measurements)
	for i := 0; i < length; i++ {
		s := strings.Split(ej.Measurements[i].Configuration.Commands[0], " ")
		if instance == s[len(s)-1] {
			// remove measurement by replacing with last one
			ej.Measurements[i] = ej.Measurements[length-1]
			// cut last measurement
			ej.Measurements = ej.Measurements[:length-1]
		}
	}
}
