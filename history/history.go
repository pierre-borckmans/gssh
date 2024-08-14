package history

import (
	"encoding/json"
	"fmt"
	"gssh/gcloud"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type Connection struct {
	Index      int
	ConfigName string
	Instance   *gcloud.Instance
	Timestamp  time.Time
}

func (c *Connection) Title() string {
	if c.Index < 10 {
		return fmt.Sprintf("[%v] %s", c.Index, c.Instance.Name)
	}
	return c.Instance.Name
}
func (c *Connection) Description() string {
	zoneSplit := strings.Split(c.Instance.Zone, "/")
	zone := zoneSplit[len(zoneSplit)-1]
	return fmt.Sprintf("%s - %s - %s", c.Timestamp.Format("02/01/2006 15:04:05"), c.ConfigName, zone)
}
func (c *Connection) FilterValue() string {
	return c.Instance.Name
}

var historyDir string

var historyFile string
var history []*Connection

func init() {
	userConfigDir, _ := os.UserHomeDir()
	historyDir = path.Join(userConfigDir, ".gssh")
	_ = os.MkdirAll(historyDir, 0755)

	// load or create the history file
	historyFile = path.Join(historyDir, "history.json")
	_, err := os.Stat(historyFile)
	if os.IsNotExist(err) {
		f, _ := os.Create(historyFile)
		defer func() {
			_ = f.Close()
		}()
		_, _ = f.WriteString("[]")
	}
}

func ListHistory() ([]*Connection, error) {
	var history []*Connection
	bytes, err := os.ReadFile(historyFile)
	_ = json.Unmarshal(bytes, &history)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.After(history[j].Timestamp)
	})
	for i, conn := range history {
		conn.Index = i
	}
	return history, err
}

func AddConnection(configName string, i *gcloud.Instance) {
	var conn *Connection
	if conn == nil {
		conn = &Connection{
			ConfigName: configName,
			Instance:   i,
			Timestamp:  time.Now(),
		}
		history = append(history, conn)
	} else {
		conn.Timestamp = time.Now()
	}
	bytes, err := json.Marshal(history)
	if err == nil {
		_ = os.WriteFile(historyFile, bytes, 0644)
	}
}

func ClearHistory() {
	history = make([]*Connection, 0)
	bytes, err := json.Marshal(history)
	if err == nil {
		_ = os.WriteFile(historyFile, bytes, 0644)
	}
}
