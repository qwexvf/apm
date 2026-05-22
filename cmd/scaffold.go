package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Scaffold a new artifact (skill, ...)",
}

var scaffoldSkillCmd = &cobra.Command{
	Use:   "skill <name>",
	Short: "Scaffold a new SKILL.md under ./skills/<name>/",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if name == "" || strings.ContainsAny(name, "/\\ ") {
			return fmt.Errorf("invalid skill name %q: must be a single path-safe segment", name)
		}

		base, err := os.Getwd()
		if err != nil {
			return err
		}
		dir := filepath.Join(base, "skills", name)
		if _, err := os.Stat(dir); err == nil {
			return fmt.Errorf("%s already exists", dir)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		file := filepath.Join(dir, "SKILL.md")
		body := fmt.Sprintf(skillTemplate, name)
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			return err
		}

		fmt.Printf("created %s\n", file)
		fmt.Println("next: edit the frontmatter description, then commit + push,")
		fmt.Printf("      then: apm add %s@<owner>/<repo>:skills/%s\n", name, name)
		return nil
	},
}

const skillTemplate = `---
name: %s
description: One-line summary of what this skill does and when it triggers.
---

# %[1]s

Describe the skill's behavior here. Keep instructions focused on what the
model should do when this skill is loaded.
`

func init() {
	scaffoldCmd.AddCommand(scaffoldSkillCmd)
}
