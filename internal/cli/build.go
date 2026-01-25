package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

const defaultConfigName = ".imageset-packer.yaml"

// CmdBuild builds one or more pack projects from a config file.
type CmdBuild struct {
	Args struct {
		Path string `positional-arg-name:"path" description:"Path to config file or directory (default: ./.imageset-packer.yaml)"`
	} `positional-args:"yes"`

	Only []string `short:"p" long:"project" description:"Build only selected project names (repeatable)" yaml:"-"`
}

// Execute runs the build command.
func (c *CmdBuild) Execute(args []string) error {
	return runBuild(c)
}

func runBuild(opts *CmdBuild) error {
	configPath, err := resolveConfigPath(opts.Args.Path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	projects, err := parsePackProjects(data)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	if len(projects) == 0 {
		return fmt.Errorf("no projects found in %q", configPath)
	}

	baseDir := filepath.Dir(configPath)
	selected, err := filterProjects(projects, opts.Only, baseDir)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no projects selected")
	}

	for _, cfg := range selected {
		if err := runPack(&cfg); err != nil {
			return err
		}
	}

	return nil
}

// resolveConfigPath resolves the path to the config file.
func resolveConfigPath(arg string) (string, error) {
	if strings.TrimSpace(arg) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get cwd: %w", err)
		}
		path := filepath.Join(cwd, defaultConfigName)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("config not found: %s", path)
		}

		return path, nil
	}

	info, err := os.Stat(arg)
	if err != nil {
		return "", fmt.Errorf("config path: %w", err)
	}

	if info.IsDir() {
		path := filepath.Join(arg, defaultConfigName)
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("config not found: %s", path)
		}
		return path, nil
	}

	return arg, nil
}

// parsePackProjects parses the pack projects from the config file.
func parsePackProjects(data []byte) ([]CmdPack, error) {
	var doc struct {
		Projects []CmdPack `yaml:"projects"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Projects) > 0 {
		return doc.Projects, nil
	}

	var list []CmdPack
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return list, nil
}

// filterProjects filters the pack projects based on the selected projects.
func filterProjects(projects []CmdPack, only []string, baseDir string) ([]CmdPack, error) {
	for i := range projects {
		if err := defaults.Set(&projects[i]); err != nil {
			return nil, fmt.Errorf("apply defaults: %w", err)
		}
		normalizeProjectPaths(&projects[i], baseDir)
	}
	if len(only) == 0 {
		return projects, nil
	}

	onlySet := make(map[string]struct{}, len(only))
	for _, name := range only {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		onlySet[name] = struct{}{}
	}
	if len(onlySet) == 0 {
		return nil, fmt.Errorf("no valid --only values")
	}

	out := make([]CmdPack, 0, len(projects))
	for _, cfg := range projects {
		effective, err := resolveProjectName(&cfg)
		if err != nil {
			return nil, err
		}
		if _, ok := onlySet[effective]; ok {
			out = append(out, cfg)
		}
	}

	return out, nil
}

// resolveProjectName resolves the project name from the config.
func resolveProjectName(cfg *CmdPack) (string, error) {
	if strings.TrimSpace(cfg.Name) != "" {
		return cfg.Name, nil
	}

	if strings.TrimSpace(cfg.Args.Input) == "" {
		return "", fmt.Errorf("project input is required when name is empty")
	}

	absInput, err := filepath.Abs(cfg.Args.Input)
	if err != nil {
		return "", fmt.Errorf("abs input: %w", err)
	}

	return filepath.Base(absInput), nil
}

// normalizeProjectPaths normalizes the project paths.
func normalizeProjectPaths(cfg *CmdPack, baseDir string) {
	cfg.Args.Input = resolveRelativePath(baseDir, cfg.Args.Input)
	cfg.Args.Output = resolveRelativePath(baseDir, cfg.Args.Output)
}

// resolveRelativePath resolves the relative path to the project.
func resolveRelativePath(baseDir, path string) string {
	if strings.TrimSpace(path) == "" {
		return path
	}

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(baseDir, path)
}
