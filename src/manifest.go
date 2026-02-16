package src

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// TargetConfig represents a single target in the manifest
type TargetConfig struct {
	Target   string `yaml:"target"`
	Partials string `yaml:"partials"`
	Comment  string `yaml:"comment"`
	Mode     string `yaml:"mode"`
	Backup   *bool  `yaml:"backup"`
}

// ManifestDefaults represents the defaults section of the manifest
type ManifestDefaults struct {
	Comment string `yaml:"comment"`
	Mode    string `yaml:"mode"`
	Backup  bool   `yaml:"backup"`
}

// Manifest represents a parsed .parts.yaml file
type Manifest struct {
	Defaults ManifestDefaults        `yaml:"defaults"`
	Targets  map[string]TargetConfig `yaml:"targets"`
}

// LoadManifest reads and validates a .parts.yaml file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest '%s': %w", path, err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest '%s': %w", path, err)
	}

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// validate checks the manifest for required fields and valid values
func (m *Manifest) validate() error {
	if len(m.Targets) == 0 {
		return fmt.Errorf("no targets defined in manifest")
	}

	validModes := map[string]bool{"merge": true, "own": true, "": true}

	for name, target := range m.Targets {
		if target.Target == "" {
			return fmt.Errorf("target '%s': missing 'target' path", name)
		}
		if target.Partials == "" {
			return fmt.Errorf("target '%s': missing 'partials' path", name)
		}
		if !validModes[target.Mode] {
			return fmt.Errorf("target '%s': invalid mode '%s' (must be 'merge' or 'own')", name, target.Mode)
		}
	}

	return nil
}

// ResolvedTarget returns a TargetConfig with defaults applied
func (m *Manifest) ResolvedTarget(name string) TargetConfig {
	target := m.Targets[name]

	// Apply defaults for empty fields
	if target.Comment == "" {
		if m.Defaults.Comment != "" {
			target.Comment = m.Defaults.Comment
		} else {
			target.Comment = "auto"
		}
	}

	if target.Mode == "" {
		if m.Defaults.Mode != "" {
			target.Mode = m.Defaults.Mode
		} else {
			target.Mode = "merge"
		}
	}

	if target.Backup == nil {
		backup := m.Defaults.Backup
		target.Backup = &backup
	}

	return target
}

// FilterTargets returns sorted target names, filtered by the given names.
// If names is nil or empty, returns all target names sorted.
// Returns an error if any requested name doesn't exist.
func (m *Manifest) FilterTargets(names []string) ([]string, error) {
	if len(names) == 0 {
		all := make([]string, 0, len(m.Targets))
		for name := range m.Targets {
			all = append(all, name)
		}
		sort.Strings(all)
		return all, nil
	}

	for _, name := range names {
		if _, exists := m.Targets[name]; !exists {
			available := make([]string, 0, len(m.Targets))
			for k := range m.Targets {
				available = append(available, k)
			}
			sort.Strings(available)
			return nil, fmt.Errorf("unknown target '%s' (available: %v)", name, available)
		}
	}

	return names, nil
}
