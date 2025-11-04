package agent

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/opencode/opencode-go/internal/config"
)

type Permission struct {
	Edit     string            `json:"edit"`
	Bash     map[string]string `json:"bash"`
	Webfetch string            `json:"webfetch,omitempty"`
}

type ModelInfo struct {
	ModelID    string `json:"modelID"`
	ProviderID string `json:"providerID"`
}

type Info struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Mode        string                 `json:"mode"`
	BuiltIn     bool                   `json:"builtIn"`
	TopP        *float64               `json:"topP,omitempty"`
	Temperature *float64               `json:"temperature,omitempty"`
	Permission  Permission             `json:"permission"`
	Model       *ModelInfo             `json:"model,omitempty"`
	Prompt      string                 `json:"prompt,omitempty"`
	Tools       map[string]bool        `json:"tools"`
	Options     map[string]interface{} `json:"options"`
}

type agentConfigEntry struct {
	Disable     bool                   `json:"disable,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Prompt      string                 `json:"prompt,omitempty"`
	Tools       map[string]bool        `json:"tools,omitempty"`
	Description string                 `json:"description,omitempty"`
	Temperature *float64               `json:"temperature,omitempty"`
	TopP        *float64               `json:"top_p,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
	Permission  map[string]interface{} `json:"permission,omitempty"`
	Extra       map[string]interface{} `json:"-"`
}

type projectConfig struct {
	Agent      map[string]agentConfigEntry `json:"agent,omitempty"`
	Permission map[string]interface{}      `json:"permission,omitempty"`
	Tools      map[string]bool             `json:"tools,omitempty"`
}

var cachedAgents []Info

func defaultPermission() Permission {
	return Permission{
		Edit: "allow",
		Bash: map[string]string{
			"*": "allow",
		},
		Webfetch: "allow",
	}
}

func planPermission() Permission {
	return Permission{
		Edit: "deny",
		Bash: map[string]string{
			"cut*":             "allow",
			"diff*":            "allow",
			"du*":              "allow",
			"file *":           "allow",
			"find * -delete*":  "ask",
			"find * -exec*":    "ask",
			"find * -fprint*":  "ask",
			"find * -fls*":     "ask",
			"find * -fprintf*": "ask",
			"find * -ok*":      "ask",
			"find *":           "allow",
			"git diff*":        "allow",
			"git log*":         "allow",
			"git show*":        "allow",
			"git status*":      "allow",
			"git branch":       "allow",
			"git branch -v":    "allow",
			"grep*":            "allow",
			"head*":            "allow",
			"less*":            "allow",
			"ls*":              "allow",
			"more*":            "allow",
			"pwd*":             "allow",
			"rg*":              "allow",
			"sort --output=*":  "ask",
			"sort -o *":        "ask",
			"sort*":            "allow",
			"stat*":            "allow",
			"tail*":            "allow",
			"tree -o *":        "ask",
			"tree*":            "allow",
			"uniq*":            "allow",
			"wc*":              "allow",
			"whereis*":         "allow",
			"which*":           "allow",
			"*":                "ask",
		},
		Webfetch: "allow",
	}
}

func List() ([]Info, error) {
	if cachedAgents != nil {
		return cachedAgents, nil
	}

	defaultTools := make(map[string]bool)
	defaultPerm := defaultPermission()

	agents := make(map[string]Info)

	// Built-in agents
	agents["general"] = Info{
		Name: "general",
		Description: "General-purpose agent for researching complex questions, searching for code, and executing multi-step tasks. " +
			"When you are searching for a keyword or file and are not confident that you will find the right match in the first few tries use this agent to perform the search for you.",
		Tools: map[string]bool{
			"todoread":  false,
			"todowrite": false,
		},
		Options:    make(map[string]interface{}),
		Permission: defaultPerm,
		Mode:       "subagent",
		BuiltIn:    true,
	}

	agents["build"] = Info{
		Name:       "build",
		Tools:      copyTools(defaultTools),
		Options:    make(map[string]interface{}),
		Permission: defaultPerm,
		Mode:       "primary",
		BuiltIn:    true,
	}

	agents["plan"] = Info{
		Name:       "plan",
		Options:    make(map[string]interface{}),
		Permission: planPermission(),
		Tools:      copyTools(defaultTools),
		Mode:       "primary",
		BuiltIn:    true,
	}

	// Load custom agents from config
	projectCfg, err := loadProjectConfig()
	if err == nil && projectCfg != nil {
		if projectCfg.Tools != nil {
			for k, v := range projectCfg.Tools {
				defaultTools[k] = v
			}
		}

		for key, value := range projectCfg.Agent {
			if value.Disable {
				delete(agents, key)
				continue
			}

			item, exists := agents[key]
			if !exists {
				item = Info{
					Name:       key,
					Mode:       "all",
					Permission: defaultPerm,
					Options:    make(map[string]interface{}),
					Tools:      make(map[string]bool),
					BuiltIn:    false,
				}
			}

			if value.Model != "" {
				// Parse model string (format: "provider:model")
				// Simplified - in production, use proper provider parsing
				item.Model = &ModelInfo{
					ProviderID: "anthropic",
					ModelID:    value.Model,
				}
			}

			if value.Prompt != "" {
				item.Prompt = value.Prompt
			}

			if value.Tools != nil {
				for k, v := range value.Tools {
					item.Tools[k] = v
				}
			}

			if value.Description != "" {
				item.Description = value.Description
			}

			if value.Temperature != nil {
				item.Temperature = value.Temperature
			}

			if value.TopP != nil {
				item.TopP = value.TopP
			}

			if value.Mode != "" {
				item.Mode = value.Mode
			}

			if value.Name != "" {
				item.Name = value.Name
			}

			// Merge default tools
			for k, v := range defaultTools {
				if _, exists := item.Tools[k]; !exists {
					item.Tools[k] = v
				}
			}

			agents[key] = item
		}
	}

	// Convert map to slice
	result := make([]Info, 0, len(agents))
	for _, agent := range agents {
		result = append(result, agent)
	}

	cachedAgents = result
	return result, nil
}

func Get(name string) (*Info, error) {
	agents, err := List()
	if err != nil {
		return nil, err
	}

	for _, agent := range agents {
		if agent.Name == name {
			return &agent, nil
		}
	}

	return nil, nil
}

func loadProjectConfig() (*projectConfig, error) {
	// Try to load opencode.json from current directory or config directory
	configPaths := []string{
		"./opencode.json",
		"./opencode.jsonc",
		filepath.Join(config.Path.Config, "opencode.json"),
		filepath.Join(config.Path.Config, "opencode.jsonc"),
	}

	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cfg projectConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}

		return &cfg, nil
	}

	return nil, nil
}

func copyTools(tools map[string]bool) map[string]bool {
	result := make(map[string]bool, len(tools))
	for k, v := range tools {
		result[k] = v
	}
	return result
}

// GetSubagentTypes returns all available subagent types for the task tool
func GetSubagentTypes() ([]string, []string, error) {
	agents, err := List()
	if err != nil {
		return nil, nil, err
	}

	types := make([]string, 0)
	descriptions := make([]string, 0)

	for _, agent := range agents {
		// Include agents with mode "subagent" or "all"
		if agent.Mode == "subagent" || agent.Mode == "all" {
			types = append(types, agent.Name)
			desc := agent.Name
			if agent.Description != "" {
				desc += ": " + agent.Description
			}
			descriptions = append(descriptions, desc)
		}
	}

	return types, descriptions, nil
}
