package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const marketplaceFile = "plugins/known_marketplaces.json"

// KnownMarketplaces is the known_marketplaces.json structure.
type KnownMarketplaces map[string]MarketplaceRecord

type MarketplaceRecord struct {
	Source          MarketplaceRecordSource `json:"source"`
	InstallLocation string                  `json:"installLocation"`
}

type MarketplaceRecordSource struct {
	Source string `json:"source"` // "github"
	Repo   string `json:"repo"`   // "org/repo"
}

// LoadKnownMarketplaces reads known_marketplaces.json from claudeDir.
func LoadKnownMarketplaces(claudeDir string) (KnownMarketplaces, error) {
	path := filepath.Join(claudeDir, marketplaceFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return KnownMarketplaces{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	m := KnownMarketplaces{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse known_marketplaces.json: %w", err)
	}
	return m, nil
}

// Save writes known_marketplaces.json atomically.
func (km KnownMarketplaces) Save(claudeDir string) error {
	path := filepath.Join(claudeDir, marketplaceFile)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(km, "", "  ")
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

// Add registers a marketplace.
func (km KnownMarketplaces) Add(id, repo, installLocation string) {
	km[id] = MarketplaceRecord{
		Source:          MarketplaceRecordSource{Source: "github", Repo: repo},
		InstallLocation: installLocation,
	}
}

// Repo returns the github repo for a marketplace ID, or "".
func (km KnownMarketplaces) Repo(id string) string {
	if r, ok := km[id]; ok {
		return r.Source.Repo
	}
	return ""
}
