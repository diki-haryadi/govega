package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/diki-haryadi/govega/api"
	"github.com/diki-haryadi/govega/env"
	"github.com/diki-haryadi/govega/util"
	"gopkg.in/yaml.v2"
)

// Getter configs getter interface
type Getter interface {
	Get(k string) interface{}
	GetString(k string) string
	GetBool(k string) bool
	GetInt(k string) int
	GetFloat64(k string) float64
	GetStringSlice(k string) []string
	GetStringMap(k string) map[string]interface{}
	GetStringMapString(k string) map[string]string
	Unmarshal(rawVal interface{}) error
}

func Load(defaultConfig map[string]interface{}, uri string) (Getter, error) {
	if uri == "" {
		readEnvVar(defaultConfig, "")
		return NewEmbedConfig(defaultConfig), nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "env":
		readEnvVar(defaultConfig, u.Query().Get("prefix"))
	case "file":
		if err := readConfigFile(defaultConfig, u); err != nil {
			return nil, err
		}
		readEnvVar(defaultConfig, u.Query().Get("prefix"))
	case "http", "https":
		if err := readRemote(defaultConfig, u); err != nil {
			return nil, err
		}
		readEnvVar(defaultConfig, u.Query().Get("prefix"))
	default:
		return nil, errors.New("unsupported scheme")
	}

	return NewEmbedConfig(defaultConfig), nil
}

func readEnvVar(defaultConfig map[string]interface{}, prefix string) {
	for k, v := range defaultConfig {
		if obj, ok := v.(map[string]interface{}); ok {
			readEnvVar(obj, prefix+k+"_")
			defaultConfig[k] = obj
			continue
		}
		if val := os.Getenv(strings.ToUpper(prefix + k)); val != "" {
			defaultConfig[k] = val
			continue
		}
	}
}

func readConfigFile(defaultConfig map[string]interface{}, uri *url.URL) error {
	path := filepath.Join(uri.Host, uri.Path)
	ext := filepath.Ext(path)
	environ := env.Get()
	if environ != "" {
		environ = "." + environ
	}
	fname := strings.TrimSuffix(path, ext) + environ + ext

	var out map[string]interface{}

	switch ext {
	case ".json":
		jsonFile, err := ioutil.ReadFile(fname)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(jsonFile, &out); err != nil {
			return err
		}
	case ".yaml":
		yamlFile, err := ioutil.ReadFile(fname)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(yamlFile, &out); err != nil {
			return err
		}
	default:
		return errors.New("unsupported file format")
	}

	for k := range defaultConfig {
		if val, ok := out[k]; ok {
			defaultConfig[k] = val
		}
	}
	return nil
}

func readRemote(defaultConfig map[string]interface{}, uri *url.URL) error {
	var out map[string]interface{}

	environ := env.Get()

	if environ != "" {
		q := uri.Query()
		q.Add("env", environ)
		uri.RawQuery = q.Encode()
	}
	if err := api.Get(uri.String()).Execute().Consume(&out); err != nil {
		return err
	}

	for k := range defaultConfig {
		if val, ok := out[k]; ok {
			defaultConfig[k] = val
		}
	}

	return nil
}

func EnvToStruct(obj interface{}) error {
	tags, err := util.ListTag(obj, "json")
	if err != nil {
		return err
	}

	conf := make(map[string]interface{})
	for _, t := range tags {
		if strings.Contains(t, ",") {
			t = strings.Split(t, ",")[0]
		}
		conf[t] = os.Getenv(strings.ToUpper(t))
	}

	return util.DecodeJSON(conf, obj)
}

func MergeConfig(obj interface{}, tag string) error {
	tags, err := util.ListTag(obj, tag)
	if err != nil {
		return err
	}

	for _, t := range tags {
		if strings.Contains(t, ",") {
			t = strings.Split(t, ",")[0]
		}

		if ev := os.Getenv(strings.ToUpper(t)); ev != "" {
			fn, err := util.FindFieldByTag(obj, tag, t)
			if err != nil {
				return err
			}

			if val, err := util.DecodeString(ev); err == nil {
				if err := util.SetValue(obj, fn, val); err != nil {
					return err
				}
			}

		}

	}

	return nil
}
