package gcloud

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
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

func (i Instance) Title() string       { return i.Name }
func (i Instance) Description() string { return path.Base(i.Zone) }
func (i Instance) FilterValue() string { return i.Name }

func ListInstances() ([]*Instance, error) {
	cacheFile := "instances_cache.json"
	if cached, err := os.ReadFile(cacheFile); err == nil {
		var instances []*Instance
		if json.Unmarshal(cached, &instances) == nil {
			return instances, nil
		}
	}

	cmd := exec.Command("gcloud", "compute", "instances", "list", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error fetching instances:", err)
		return nil, err
	}

	var rawInstances []map[string]interface{}
	if err = json.Unmarshal(output, &rawInstances); err != nil {
		fmt.Println("Error parsing instance list:", err)
		return nil, err
	}

	instances := make([]*Instance, len(rawInstances))
	for i, raw := range rawInstances {
		instances[i] = &Instance{
			Name:   raw["name"].(string),
			Zone:   raw["zone"].(string),
			Status: InstanceStatus(raw["status"].(string)),
		}
	}

	cacheData, _ := json.Marshal(instances)
	if err := os.WriteFile(cacheFile, cacheData, 0644); err != nil {
		fmt.Println("Error writing cache file:", err)
	}

	return instances, nil
}

func (i Instance) SSH() error {
	zone := strings.Split(i.Zone, "/")
	zoneFlag := "--zone=" + zone[len(zone)-1]
	cmd := exec.Command("gcloud", "compute", "ssh", fmt.Sprintf("%s@%s", "conductor", i.Name), zoneFlag)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
