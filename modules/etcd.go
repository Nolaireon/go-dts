package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"go.etcd.io/etcd/client"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func FetchApps(config *Etcd, url, hostName string) (err error) {
	uri := fmt.Sprintf("%s/v2/keys/ps/hosts/%s/apps?recursive=true", url, hostName)
	resp, err := http.Get(uri)
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

func (config *Etcd) ExtractAppByInstance(instance string, app *App) (ok bool, err error) {
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
				case "app_dir":
					app.AppDir = v.Value
				case "appl_id":
					app.ApplId = v.Value
				case "application_name":
					app.ApplicationName = v.Value
				case "current_version":
					app.CurrentVersion = v.Value
				case "history":
					app.History = v.Value
				case "info_json":
					app.InfoJson = v.Value
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

func (app *App) WriteApp(uri string, kApi client.KeysAPI) error {
	v := reflect.ValueOf(app).Elem()
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Tag.Get("json")
		switch v.Field(i).Interface().(type) {
		case string:
			if v.Field(i).Len() > 0 {
				if resp, err := kApi.Set(context.Background(), uri+key, v.Field(i).String(), nil); err != nil {
					log.Printf("key update err: %s\n", uri+key)
					return err
				} else {
					log.Printf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Interface())
				}
			}
		case EmonJson:
			if cmp.Equal(&app.EmonJson, &EmonJson{}) {
				continue
			}

			buf, err := json.MarshalIndent(app.EmonJson, "", "    ")
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}

			resp, err := kApi.Set(context.Background(), uri+key, string(buf), nil)
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}
			log.Printf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Interface())
		case DtsSettings:
			if cmp.Equal(&app.DtsSettings, &DtsSettings{}) {
				continue
			}

			buf, err := json.MarshalIndent(app.DtsSettings, "", "    ")
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}

			resp, err := kApi.Set(context.Background(), uri+key, string(buf), nil)
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}
			log.Printf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Interface())
		default:
			log.Println("undefined type:", reflect.TypeOf(v.Field(i).Interface()))
		}
	}
	return nil
}

func (ds *DtsSettings) SetDtsSettings(appName, workTree, dtsDir, instance string) {
	gitDir := dtsDir + "/" + instance
	if len(ds.AppList) == 0 {
		ds.AppList = map[string]*Instance{}
	}
	ds.AppList[instance] = &Instance{
		AppName:  appName,
		WorkTree: workTree,
		GitDir:   gitDir,
		Enabled:  "True",
		LockFile: gitDir + "/" + ".lock",
	}
	ds.Updated = time.Now().Format(time.RFC3339)
}

func (ej *EmonJson) SetEmonJson(dtsId, gitDir, instance string) {
	ej.ApplId = dtsId
	ej.Description = "DTS"
	ej.Measurements = append(ej.Measurements, Measurement{
		Name: "data-tracking-system",
		Configuration: Configuration{
			Commands:   []string{gitDir + "/dts_agent -a status -i " + instance},
			DataFormat: "influx",
			Interval:   "5m",
			Timeout:    "30s",
			Type:       "exec",
		},
	})
	ej.Product = "DTS"
}

func (app *App) SetDtsApp(dtsId, name, stand string, ds *DtsSettings, ej *EmonJson) {
	app.ApplId = dtsId
	app.ApplicationName = name
	app.DtsSettings = *ds
	app.EmonJson = *ej
	app.Stand = stand
}
