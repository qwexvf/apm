package claude

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
)

// Settings represents the relevant parts of ~/.claude/settings.json.
// Unknown keys are preserved via rawExtra.
type Settings struct {
	EnabledPlugins       map[string]bool        `json:"enabledPlugins,omitempty"`
	ExtraKnownMarketplaces []MarketplaceEntry   `json:"extraKnownMarketplaces,omitempty"`
	rest                 map[string]json.RawMessage
}

type MarketplaceEntry struct {
	Source          string `json:"source"`
	InstallLocation string `json:"installLocation,omitempty"`
}

// LoadSettings reads settings.json from claudeDir.
func LoadSettings(claudeDir string) (*Settings, error) {
	path := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{
				EnabledPlugins: map[string]bool{},
				rest:           map[string]json.RawMessage{},
			}, nil
		}
		return nil, fmt.Errorf("read settings.json: %w", err)
	}

	// unmarshal into raw map to preserve all keys
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse settings.json: %w", err)
	}

	s := &Settings{
		EnabledPlugins: map[string]bool{},
		rest:           map[string]json.RawMessage{},
	}

	if v, ok := raw["enabledPlugins"]; ok {
		if err := json.Unmarshal(v, &s.EnabledPlugins); err != nil {
			return nil, fmt.Errorf("parse enabledPlugins: %w", err)
		}
		delete(raw, "enabledPlugins")
	}
	if v, ok := raw["extraKnownMarketplaces"]; ok {
		if err := json.Unmarshal(v, &s.ExtraKnownMarketplaces); err != nil {
			return nil, fmt.Errorf("parse extraKnownMarketplaces: %w", err)
		}
		delete(raw, "extraKnownMarketplaces")
	}

	s.rest = raw
	return s, nil
}

// Save writes settings.json atomically, preserving unknown keys.
func (s *Settings) Save(claudeDir string) error {
	path := filepath.Join(claudeDir, "settings.json")

	// rebuild full map
	out := make(map[string]json.RawMessage, len(s.rest)+2)
	maps.Copy(out, s.rest)

	if len(s.EnabledPlugins) > 0 {
		b, err := json.Marshal(s.EnabledPlugins)
		if err != nil {
			return err
		}
		out["enabledPlugins"] = b
	}
	if len(s.ExtraKnownMarketplaces) > 0 {
		b, err := json.Marshal(s.ExtraKnownMarketplaces)
		if err != nil {
			return err
		}
		out["extraKnownMarketplaces"] = b
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// EnablePlugin sets the plugin enabled state.
func (s *Settings) EnablePlugin(id string, enabled bool) {
	if s.EnabledPlugins == nil {
		s.EnabledPlugins = map[string]bool{}
	}
	s.EnabledPlugins[id] = enabled
}

// DisablePlugin removes the plugin entry entirely.
func (s *Settings) DisablePlugin(id string) {
	delete(s.EnabledPlugins, id)
}

// AddMarketplace adds a marketplace URL if not already present.
func (s *Settings) AddMarketplace(source, installLocation string) {
	for _, e := range s.ExtraKnownMarketplaces {
		if e.Source == source {
			return
		}
	}
	s.ExtraKnownMarketplaces = append(s.ExtraKnownMarketplaces, MarketplaceEntry{
		Source:          source,
		InstallLocation: installLocation,
	})
}
