package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path"
)

type SSHConfig struct {
	UserName string `toml:"user_name"`
}

type InstancesConfig struct {
	Exclusions []string `toml:"exclusions"`
}

type Configuration struct {
	SSH       SSHConfig       `toml:"ssh"`
	Instances InstancesConfig `toml:"instances"`
}

var Config Configuration

var defaultConfigStr = `
[ssh]
user_name = "conductor"

[instances]
exclusions = ["gke-"]
`

func init() {
	Config = Configuration{}
	userConfigDir, _ := os.UserHomeDir()
	configDir := path.Join(userConfigDir, ".gssh")
	configFilepath := path.Join(configDir, "config.toml")

	if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0755)
		f, _ := os.Create(configFilepath)
		defer func() {
			_ = f.Close()
		}()
		_, _ = f.WriteString(defaultConfigStr)
	}

	if _, err := toml.DecodeFile(configFilepath, &Config); err != nil {
		log.Fatal("Error decoding config file:", err)
	}
}
