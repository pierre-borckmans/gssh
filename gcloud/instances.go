package gcloud

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"gssh/config"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type InstanceStatus string

var (
	InstanceStatusRunning  = InstanceStatus("RUNNING")
	InstanceStatusStopped  = InstanceStatus("STOPPED")
	InstanceStatusTerminal = InstanceStatus("TERMINATED")
)

type Instance struct {
	Name   string
	Zone   string
	Status InstanceStatus
}

var _ list.Item = &Instance{}

func (i *Instance) Title() string       { return i.Name }
func (i *Instance) Description() string { return path.Base(i.Zone) }
func (i *Instance) FilterValue() string {
	return i.Name
}

var cacheDir string
var exclusions []string

func init() {
	userConfigDir, _ := os.UserHomeDir()
	cacheDir = path.Join(userConfigDir, ".gssh")
	_ = os.MkdirAll(cacheDir, 0755)

	validExclusions := make([]string, 0)
	for _, ex := range config.Config.Instances.Exclusions {
		if strings.Trim(ex, " ") != "" {
			validExclusions = append(validExclusions, ex)
		}
	}
	exclusions = validExclusions
}

func ListInstances(configName string, clearCache bool) ([]*Instance, *time.Time, error) {
	var instances []*Instance
	var lastUpdate = time.Now()
	foundCache := false

	cacheFile := path.Join(cacheDir, fmt.Sprintf("instances_cache_%v.json", configName))
	if !clearCache {
		if cached, err := os.ReadFile(cacheFile); err == nil {
			_ = json.Unmarshal(cached, &instances)
			foundCache = true
			s, err := os.Stat(cacheFile)
			if err == nil {
				lastUpdate = s.ModTime()
			}
		}
	} else {
		_ = os.Remove(cacheFile)
	}

	if instances == nil || len(instances) == 0 {
		cmd := exec.Command("gcloud", "compute", "instances", "list", "--format=json", "--configuration", configName)
		output, err := cmd.Output()
		if err != nil {
			return nil, nil, err
		}

		var rawInstances []map[string]interface{}
		if err = json.Unmarshal(output, &rawInstances); err != nil {
			return nil, nil, err
		}

		instances = make([]*Instance, len(rawInstances))
		for i, raw := range rawInstances {
			instances[i] = &Instance{
				Name:   raw["name"].(string),
				Zone:   raw["zone"].(string),
				Status: InstanceStatus(raw["status"].(string)),
			}
		}
	}

	filteredInstances := make([]*Instance, 0)
	for _, inst := range instances {
		excluded := false
		for _, ex := range exclusions {
			if strings.Contains(inst.Name, ex) {
				excluded = true
				continue
			}
		}
		if excluded {
			continue
		}
		filteredInstances = append(filteredInstances, inst)
	}

	if !foundCache {
		cacheData, _ := json.Marshal(filteredInstances)
		if err := os.WriteFile(cacheFile, cacheData, 0644); err != nil {
		}
	}

	return filteredInstances, &lastUpdate, nil
}

func (i *Instance) SSH(configName string) error {
	zone := strings.Split(i.Zone, "/")
	zoneFlag := "--zone=" + zone[len(zone)-1]
	cmd := exec.Command("gcloud", "compute", "ssh", "--configuration", configName, fmt.Sprintf("%s@%s", config.Config.SSH.UserName, i.Name), zoneFlag)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
