package env

import "os"

const (
	EnvDev  = "development"
	EnvStag = "staging"
	EnvProd = "production"
)

var (
	env string
)

func init() {
	// read variabel global env engine
	env = os.Getenv("SCENV")
	if env == "" {
		env = EnvDev
	}
}

func Get() string {
	return env
}

func IsDev() bool {
	return EnvDev == env
}

func IsStag() bool {
	return EnvStag == env
}

func IsProd() bool {
	return EnvProd == env
}
