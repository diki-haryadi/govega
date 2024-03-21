package config

import (
	"fmt"
	"strconv"

	"github.com/dikiharyadi19/govegapunk/util"
	"github.com/mitchellh/mapstructure"
)

type EmbedConfig struct {
	config interface{}
}

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

func NewEmbedConfig(config interface{}) *EmbedConfig {
	if config == nil {
		return nil
	}
	return &EmbedConfig{
		config: config,
	}
}

// Get object
func (e *EmbedConfig) Get(k string) interface{} {
	if k == "" {
		return e.config
	}
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return nil
	}
	return v
}

// GetString get string value
func (e *EmbedConfig) GetString(k string) string {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// GetBool get bool value
func (e *EmbedConfig) GetBool(k string) bool {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return false
	}
	switch v := v.(type) {
	case bool:
		return v
	case string:
		b, _ := strconv.ParseBool(v)
		return b
	default:
		return false
	}
}

// GetInt get int value
func (e *EmbedConfig) GetInt(k string) int {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return 0
	}
	switch v := v.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

// GetFloat64 get float64 value
func (e *EmbedConfig) GetFloat64(k string) float64 {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return 0
	}
	switch v := v.(type) {
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

// GetStringSlice get string slice value
func (e *EmbedConfig) GetStringSlice(k string) []string {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return nil
	}
	var s []string
	if err := mapstructure.Decode(v, &s); err != nil {
		return nil
	}
	return s
}

// GetStringMapString get map string string value
func (e *EmbedConfig) GetStringMapString(k string) map[string]string {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return nil
	}
	var s map[string]string
	if err := mapstructure.Decode(v, &s); err != nil {
		return nil
	}
	return s
}

// GetStringMap get map string interface value
func (e *EmbedConfig) GetStringMap(k string) map[string]interface{} {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return nil
	}
	var s map[string]interface{}
	if err := mapstructure.Decode(v, &s); err != nil {
		return nil
	}
	return s
}

// Unmarshal unmarshal
func (e *EmbedConfig) Unmarshal(rawVal interface{}) error {
	return util.DecodeJSON(e.config, rawVal)
}

// GetConfig get config
func (e *EmbedConfig) GetConfig(k string) Getter {
	v, ok := util.Lookup(k, e.config)
	if !ok {
		return nil
	}

	return &EmbedConfig{
		config: v,
	}
}
