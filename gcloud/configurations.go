package gcloud

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Configuration struct {
	Name    string `json:"name"`
	Account string `json:"account"`
	Project string `json:"project"`
	Active  bool   `json:"is_active"`
}

func (c Configuration) Title() string { return c.Name }
func (c Configuration) Description() string {
	return fmt.Sprintf("Account: %s, Project: %s", c.Account, c.Project)
}
func (c Configuration) FilterValue() string { return c.Name }

func (c *Configuration) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.Name = raw["name"].(string)
	c.Active = raw["is_active"].(bool)

	properties := raw["properties"].(map[string]interface{})
	core := properties["core"].(map[string]interface{})
	c.Account = core["account"].(string)
	c.Project = core["project"].(string)
	return nil
}

func ListConfigurations() ([]*Configuration, error) {
	cmd := exec.Command("gcloud", "config", "configurations", "list", "--format=json")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error fetching configurations:", err)
		return nil, err
	}
	var configurations []*Configuration
	if err := json.Unmarshal(output, &configurations); err != nil {
		fmt.Println("Error parsing configurations:", err)
		return nil, err
	}
	return configurations, nil
}
