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

func init() {
	Config = Configuration{}
	userConfigDir, _ := os.UserHomeDir()
	configFilepath := path.Join(userConfigDir, ".gssh", "config.toml")

	if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
		f, _ := os.Create(configFilepath)
		defer func() {
			_ = f.Close()
		}()
		_, _ = f.WriteString(`
[ssh]
user_name = "conductor"

[instances]
exclusions = ["gke-",""]
`)
	}

	if _, err := toml.DecodeFile(configFilepath, &Config); err != nil {
		log.Fatal("Error decoding config file:", err)
	}
}
