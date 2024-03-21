package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/dikiharyadi19/govegapunk/env"
)

func ReadConfig(cfg interface{}, path, module string) error {

	e := env.Get()

	var file string
	filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		file = filepath.Ext(info.Name())
		return nil
	})

	switch file {
	case ".json":
		filename := fmt.Sprintf("%s/%s.%s.json", path, module, e)
		jsonFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		return json.Unmarshal(jsonFile, cfg)
	default:
		filename := fmt.Sprintf("%s/%s.%s.yaml", path, module, e)
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		return yaml.Unmarshal(yamlFile, cfg)
	}
}
