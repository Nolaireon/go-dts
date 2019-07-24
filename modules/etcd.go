package modules

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/client"
	"log"
	"reflect"
	"strings"
)

func Fetch(data []byte, config *Etcd) error {
	err := json.Unmarshal(data, &config)
	return err
}

func (config Etcd) Extract(instance string, app *App) error {
	for _, value := range config.Node.Nodes {
		key := strings.Split(value.Key, "/")
		idDotHash := strings.Split(key[len(key)-1], ".")
		if idDotHash[1] == instance {
			for _, v := range value.Nodes {
				switch k := strings.Split(v.Key, "/"); k[len(k)-1] {
				case "dts_settings":
					tempV := strings.ReplaceAll(v.Value, "\n", "")
					tempV = strings.ReplaceAll(v.Value, "    ", "")
					if err := json.Unmarshal([]byte(tempV), &app.DtsSettings); err != nil {
						return err
					}
				case "emon_json":
					tempV := strings.ReplaceAll(v.Value, "\n", "")
					tempV = strings.ReplaceAll(v.Value, "    ", "")
					if err := json.Unmarshal([]byte(tempV), &app.EmonJson); err != nil {
						return err
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
		}
	}
	return nil
}

func (app App) WriteApp(uri string, kapi client.KeysAPI) error {
	v := reflect.ValueOf(&app).Elem()
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Tag.Get("json")
		switch v.Field(i).Interface().(type) {
		case string:
			if v.Field(i).Len() > 0 {
				if resp, err := kapi.Set(context.Background(), uri+key, v.Field(i).String(), nil); err != nil {
					log.Printf("key update err: %s\n", uri+key)
					return err
				} else {
					log.Printf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Interface())
				}
			}
		case EmonJson:
			if reflect.DeepEqual(&app.EmonJson, &EmonJson{}) {
				continue
			}

			buf, err := json.MarshalIndent(app.EmonJson, "", "    ")
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}

			resp, err := kapi.Set(context.Background(), uri+key, string(buf), nil)
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}
			log.Printf("key %s: %s=%v\n", resp.Action, uri+key, v.Field(i).Interface())
		case DtsSettings:
			if reflect.DeepEqual(&app.DtsSettings, &DtsSettings{}) {
				continue
			}

			buf, err := json.MarshalIndent(app.DtsSettings, "", "    ")
			if err != nil {
				log.Printf("key update err: %s\n", uri+key)
				return err
			}

			resp, err := kapi.Set(context.Background(), uri+key, string(buf), nil)
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
