package config

import (
	"testing"
	"time"
)

type MainConfig struct {
	Server   ServerConfig `yaml:"Server"`
	Database DBConfig     `yaml:"Database"`
	//Redis           []RedisConfig           `yaml:"Redis"`
	ConfigBasicAuth ConfigBasicAuthConfig `yaml:"ConfigBasicAuth"`
	// Tracer          tracer.Config
}

type (
	ServerConfig struct {
		Port            string        `yaml:"Port"`
		BasePath        string        `yaml:"BasePath"`
		GracefulTimeout time.Duration `yaml:"GracefulTimeout"`
		ReadTimeout     time.Duration `yaml:"ReadTimeout"`
		WriteTimeout    time.Duration `yaml:"WriteTimeout"`
		APITimeout      int           `yaml:"APITimeout"`
	}

	DBConfig struct {
		SlaveDSN      string `yaml:"SlaveDSN"`
		MasterDSN     string `yaml:"MasterDSN"`
		RetryInterval int    `yaml:"RetryInterval"`
		MaxIdleConn   int    `yaml:"MaxIdleConn"`
		MaxConn       int    `yaml:"MaxConn"`
	}

	/*RedisConfig struct {
		Connection string        `yaml:"Connection"`
		Password   string        `yaml:"Password"`
		Timeout    time.Duration `yaml:"Timeout"`
		MaxIdle    int           `yaml:"MaxIdle"`
	}*/
	ConfigBasicAuthConfig struct {
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
	}
)

type Config struct {
	BranchCollection        string `json:"branch_collection"`
	AddressMapCollection    string `json:"address_map_collection"`
	PickupRequestCollection string `json:"pickup_request_collection"`
	WorkOrderCollection     string `json:"work_order_collection"`
	// Tracer          tracer.Config
}

func TestConfigJson(t *testing.T) {
	m := &Config{}
	err := ReadModuleConfig(&m, "file/json", "configs")
	if err != nil {
		t.Error(err)
	}
}

func TestConfigYaml(t *testing.T) {
	m := &MainConfig{}
	err := ReadModuleConfig(&m, "file/yaml", "configs")
	if err != nil {
		t.Error(err)
	}
}
