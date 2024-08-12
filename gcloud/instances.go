package gcloud

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type InstanceStatus string

var (
	InstanceStatusRunning = InstanceStatus("RUNNING")
	//InstanceStatusStopped  = InstanceStatus("STOPPED")
	//InstanceStatusTerminal = InstanceStatus("TERMINATED")
)

type Instance struct {
	Name   string
	Zone   string
	Status InstanceStatus
}

type Connection struct {
	ConfigName string
	Instance   *Instance
	Timestamp  time.Time
}

var _ list.Item = &Instance{}

func (i *Instance) Title() string       { return i.Name }
func (i *Instance) Description() string { return path.Base(i.Zone) }
func (i *Instance) FilterValue() string {
	return i.Name
}

var cacheDir string
var historyFile string
var history []*Connection
var exclusionsFile string
var exclusions []string

func init() {
	userConfigDir, _ := os.UserHomeDir()
	cacheDir = path.Join(userConfigDir, ".gssh")
	_ = os.MkdirAll(cacheDir, 0755)

	// load or create the history file
	historyFile = path.Join(cacheDir, "history.json")
	_, err := os.Stat(historyFile)
	if os.IsNotExist(err) {
		f, _ := os.Create(historyFile)
		defer func() {
			_ = f.Close()
		}()
		_, _ = f.WriteString("[]")
	}
	history = make([]*Connection, 0)
	bytes, err := os.ReadFile(historyFile)
	_ = json.Unmarshal(bytes, &history)

	// load or create the exclusions file
	exclusionsFile = path.Join(cacheDir, "exclusions")
	_, err = os.Stat(exclusionsFile)
	if os.IsNotExist(err) {
		f, _ := os.Create(exclusionsFile)
		defer func() {
			_ = f.Close()
		}()
		_, _ = f.WriteString("")
	}
	bytes, err = os.ReadFile(exclusionsFile)
	if err == nil {
		exclusions = strings.Split(string(bytes), "\n")
	}
	validExclusions := make([]string, 0)
	for _, ex := range exclusions {
		if strings.Trim(ex, " ") != "" {
			validExclusions = append(validExclusions, ex)
		}
	}
	exclusions = validExclusions
}

func ListInstances(configName string, clearCache bool) ([]*Instance, error) {
	var instances []*Instance

	cacheFile := path.Join(cacheDir, fmt.Sprintf("instances_cache_%v.json", configName))
	if !clearCache {
		if cached, err := os.ReadFile(cacheFile); err == nil {
			_ = json.Unmarshal(cached, &instances)
		}
	} else {
		_ = os.Remove(cacheFile)
	}

	if instances == nil || len(instances) == 0 {
		cmd := exec.Command("gcloud", "compute", "instances", "list", "--format=json", "--configuration", configName)
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		var rawInstances []map[string]interface{}
		if err = json.Unmarshal(output, &rawInstances); err != nil {
			return nil, err
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

	cacheData, _ := json.Marshal(filteredInstances)
	if err := os.WriteFile(cacheFile, cacheData, 0644); err != nil {
	}

	return filteredInstances, nil
}

func (i *Instance) SSH(configName string) error {
	zone := strings.Split(i.Zone, "/")
	zoneFlag := "--zone=" + zone[len(zone)-1]
	cmd := exec.Command("gcloud", "compute", "ssh", "--configuration", configName, fmt.Sprintf("%s@%s", "conductor", i.Name), zoneFlag)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// add the connection to the history
	conn := &Connection{
		ConfigName: configName,
		Instance:   i,
		Timestamp:  time.Now(),
	}
	history = append(history, conn)
	bytes, err := json.Marshal(history)
	if err == nil {
		_ = os.WriteFile(historyFile, bytes, 0644)
	}

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
