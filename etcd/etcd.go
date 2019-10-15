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

// CollectApps returns a list of apps obtained from the registry host, except apps from config excluded_apps.yml
func (config *Etcd) CollectApps(excludedApps []string) (apps [][]string) {
	for _, value := range config.Node.Nodes {
		parts := strings.Split(value.Key, "/")
		idDotHash := strings.Split(parts[len(parts)-1], ".")

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
	replacer := strings.NewReplacer("\n", "", "    ", "", "\t", "")
	for _, value := range config.Node.Nodes {
		// slice of key path, example: /ps/hosts/vlg/vlg-lhrs-app1d/apps/5118.3049088120
		keyParts := strings.Split(value.Key, "/")
		// example: key = ["5118.3049088120", "17013.3677141858", ...]
		key := keyParts[len(keyParts)-1]
		// example: idDotHash = ["5118", "3049088120"]
		idDotHash := strings.Split(key, ".")
		if idDotHash[1] == instance {
			for _, v := range value.Nodes {
				// slice of key path, example: /ps/hosts/vlg/vlg-lhrs-app1d/apps/5118.3049088120/appl_id
				kParts := strings.Split(v.Key, "/")
				// example: k = ["appl_id", "application_name", ...]
				switch k := kParts[len(kParts)-1]; k {
				case "dts_settings":
					if err = json.Unmarshal([]byte(replacer.Replace(v.Value)), &app.DtsSettings); err != nil {
						return
					}
				case "emon_json":
					if err = json.Unmarshal([]byte(replacer.Replace(v.Value)), &app.EmonJson); err != nil {
						return
					}
				case "app_dir":
					app.AppDir = v.Value
				case "appl_id":
					app.ApplId = v.Value
				case "application_name":
					app.ApplicationName = v.Value
				case "current_version":
					app.CurrentVersion = v.Value
				case "instance":
					app.Instance = v.Value
				case "pid_file":
					app.PidFile = v.Value
				case "product_name":
					app.ProductName = v.Value
				case "stand":
					app.Stand = v.Value
				}
			}
			ok = true
		}
	}

	return
}

func (app *App) Push(uri string) (updatedKeys []string, err error) {
	v := reflect.ValueOf(app).Elem()
	resp := &client.Response{}
	for i := 0; i < v.NumField(); i++ {
		// get json key
		key := v.Type().Field(i).Tag.Get("json")
		key = strings.Split(key, ",")[0]
		switch v.Field(i).Interface().(type) {
		case string:
			if v.Field(i).Len() > 0 {
				resp, err = kApi.Set(context.Background(), uri+key, v.Field(i).String(), nil)
				if err != nil {
					return
				}

				updatedKeys = append(updatedKeys, fmt.Sprintf("key %s: %s=%s\n", resp.Action, uri+key, v.Field(i).String()))
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

			updatedKeys = append(updatedKeys, fmt.Sprintf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Elem().Interface()))
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

			updatedKeys = append(updatedKeys, fmt.Sprintf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Elem().Interface()))
		}
	}
	return
}

// SetDtsSettings set or update dts_settings struct
func (ds *DtsSettings) SetDtsSettings(appDir, appName, workTree, dtsDir, instance string) {
	gitDir := filepath.Join(dtsDir, instance)
	if ds.AppList == nil {
		ds.AppList = map[string]*Instance{}
	}

	ds.AppList[instance] = &Instance{
		AppDir:   appDir,
		AppName:  appName,
		WorkTree: workTree,
		GitDir:   gitDir,
		Enabled:  true,
	}
	ds.Updated = time.Now().Format(time.RFC3339)
}

// SetEmonJson set or update emon_json struct
func (ej *EmonJson) SetEmonJson(dtsId int, dtsAppName, gitDir, instance string) {
	dtsApp := filepath.Join(gitDir, strings.ToLower(dtsAppName))
	command := fmt.Sprintf("%s -a status -i %s", dtsApp, instance)
	ej.ApplId = dtsId
	ej.Description = "data-tracking-system"
	ej.Measurements = append(ej.Measurements, &Measurement{
		Name: "data-tracking-system",
		Configuration: &Configuration{
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
			// don't remove return, otherwise will panic with index out of range after cutting slice
			return
		}
	}
}
