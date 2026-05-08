package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	mktpkg "github.com/qwexvf/apm/internal/marketplace"
)

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Manage plugin marketplaces",
}

var marketplaceAddCmd = &cobra.Command{
	Use:   "add <id> <github:owner/repo>",
	Short: "Register a new marketplace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		rawSrc := args[1]

		// parse "github:owner/repo" or plain "owner/repo"
		repo := strings.TrimPrefix(rawSrc, "github:")
		if !strings.Contains(repo, "/") {
			return fmt.Errorf("invalid source %q: expected github:owner/repo", rawSrc)
		}

		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			m = config.NewManifest()
			m.PluginManager.Scope = resolveScope()
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return err
		}

		installLocation := filepath.Join(claude.MarketplaceCacheDir(claudeDir), id)
		km.Add(id, repo, installLocation)
		if err := km.Save(claudeDir); err != nil {
			return err
		}

		// also add to manifest
		m.Marketplaces[id] = config.MarketplaceSource{Source: "github", Repo: repo}
		if err := m.Save(dir); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		idx := mktpkg.New(id, repo, installLocation)
		fmt.Printf("cloning %s from github.com/%s...\n", id, repo)
		if _, err := idx.EnsureCloned(ctx); err != nil {
			return err
		}

		fmt.Printf("✓ marketplace %q registered\n", id)
		return nil
	},
}

var marketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered marketplaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			m = config.NewManifest()
			m.PluginManager.Scope = resolveScope()
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return err
		}

		if len(km) == 0 {
			fmt.Println("no marketplaces registered")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tREPO\tLOCATION")
		for id, rec := range km {
			fmt.Fprintf(w, "%s\tgithub.com/%s\t%s\n", id, rec.Source.Repo, rec.InstallLocation)
		}
		return w.Flush()
	},
}

var marketplaceUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Pull latest marketplace indexes",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			m = config.NewManifest()
			m.PluginManager.Scope = resolveScope()
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		type result struct {
			id        string
			localPath string
			oldSHA    string
			newSHA    string
			cloned    bool
			err       error
		}

		mktplaceDir := claude.MarketplaceCacheDir(claudeDir)

		// collect targets
		type target struct {
			id        string
			localPath string
			repo      string
		}
		var targets []target
		for id, rec := range km {
			if len(args) > 0 && args[0] != id {
				continue
			}
			localPath := rec.InstallLocation
			if localPath == "" {
				localPath = filepath.Join(mktplaceDir, id)
			}
			targets = append(targets, target{id: id, localPath: localPath, repo: rec.Source.Repo})
		}

		if len(targets) == 0 {
			fmt.Println("no marketplaces to update")
			return km.Save(claudeDir)
		}

		// spinner while updates run in parallel
		results := make(chan result, len(targets))
		var wg sync.WaitGroup
		for _, t := range targets {
			wg.Add(1)
			go func(t target) {
				defer wg.Done()
				idx := mktpkg.New(t.id, t.repo, t.localPath)
				ur, err := idx.EnsureCloned(ctx)
				if err != nil {
					results <- result{id: t.id, localPath: t.localPath, err: err}
					return
				}
				results <- result{
					id:        t.id,
					localPath: ur.LocalPath,
					oldSHA:    ur.OldSHA,
					newSHA:    ur.NewSHA,
					cloned:    ur.Cloned,
				}
			}(t)
		}
		go func() {
			wg.Wait()
			close(results)
		}()

		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinDone := make(chan struct{})
		go func() {
			i := 0
			for {
				select {
				case <-spinDone:
					fmt.Print("\r\033[K")
					return
				case <-time.After(80 * time.Millisecond):
					fmt.Printf("\r%s updating marketplaces...", spinner[i%len(spinner)])
					i++
				}
			}
		}()

		var mu sync.Mutex
		var collected []result
		for r := range results {
			mu.Lock()
			collected = append(collected, r)
			mu.Unlock()
		}

		close(spinDone)
		time.Sleep(90 * time.Millisecond) // let spinner goroutine clear the line

		for _, r := range collected {
			if r.err != nil {
				fmt.Printf("  ✗ %s: %v\n", r.id, r.err)
				continue
			}
			if r.cloned {
				fmt.Printf("  ✓ %s  cloned %s  %s\n", r.id, r.newSHA, r.localPath)
			} else if r.oldSHA != r.newSHA && r.oldSHA != "" {
				fmt.Printf("  ✓ %s  %s → %s  %s\n", r.id, r.oldSHA, r.newSHA, r.localPath)
			} else {
				fmt.Printf("  ✓ %s  up to date (%s)  %s\n", r.id, r.newSHA, r.localPath)
			}
		}

		return km.Save(claudeDir)
	},
}

func init() {
	marketplaceCmd.AddCommand(marketplaceAddCmd, marketplaceListCmd, marketplaceUpdateCmd)
}
