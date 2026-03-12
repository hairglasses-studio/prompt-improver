package enhancer

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds per-project prompt enhancement configuration.
// Loaded from .prompt-improver.yaml in the project directory.
type Config struct {
	// Preamble is always-injected context added before the enhanced prompt
	Preamble string `yaml:"preamble"`

	// Rules are pattern-matched augmentations
	Rules []Rule `yaml:"rules"`

	// BlockPatterns are regexes that cause the prompt to be blocked (exit 2)
	BlockPatterns []string `yaml:"block_patterns"`

	// DisabledStages allows disabling specific pipeline stages
	DisabledStages []string `yaml:"disabled_stages"`

	// DefaultTaskType overrides auto-detection
	DefaultTaskType string `yaml:"default_task_type"`

	// DefaultEffort overrides auto-detection of effort level (low, medium, high)
	DefaultEffort string `yaml:"default_effort"`

	// Hook holds configuration specific to the UserPromptSubmit hook mode
	Hook HookConfig `yaml:"hook"`
}

// HookConfig holds settings for the Claude Code UserPromptSubmit hook.
type HookConfig struct {
	// SkipScoreThreshold skips enhancement if the prompt already scores >= this (default 75, 0 = always enhance)
	SkipScoreThreshold int `yaml:"skip_score_threshold"`

	// MinWordCount skips prompts shorter than this (default 5)
	MinWordCount int `yaml:"min_word_count"`

	// SkipPatterns are additional regex patterns that cause the hook to skip enhancement
	SkipPatterns []string `yaml:"skip_patterns"`
}

// Rule is a pattern-matched augmentation rule
type Rule struct {
	Match   string `yaml:"match"`   // regex pattern on the prompt
	Prepend string `yaml:"prepend"` // context to add before the prompt
	Append  string `yaml:"append"`  // context to add after the prompt
}

// LoadConfig loads configuration from .prompt-improver.yaml in the given directory.
// Returns a zero Config if the file does not exist.
func LoadConfig(dir string) Config {
	var cfg Config

	paths := []string{
		filepath.Join(dir, ".prompt-improver.yaml"),
		filepath.Join(dir, ".prompt-improver.yml"),
		filepath.Join(dir, ".claude", "prompt-improver.yaml"),
		filepath.Join(dir, ".claude", "prompt-improver.yml"),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue
		}
		return cfg
	}

	return cfg
}

// IsStageDisabled checks if a pipeline stage is disabled in config
func (c Config) IsStageDisabled(stage string) bool {
	for _, s := range c.DisabledStages {
		if strings.EqualFold(s, stage) {
			return true
		}
	}
	return false
}

// ApplyRules applies matching rules to the prompt, returning modified text
func (c Config) ApplyRules(text string) (string, []string) {
	if len(c.Rules) == 0 {
		return text, nil
	}

	var improvements []string
	for _, rule := range c.Rules {
		if rule.Match == "" {
			continue
		}
		lower := strings.ToLower(text)
		matchLower := strings.ToLower(rule.Match)

		// Simple substring match (not regex for safety)
		if !strings.Contains(lower, matchLower) {
			continue
		}

		if rule.Prepend != "" {
			text = rule.Prepend + "\n\n" + text
			improvements = append(improvements, "Applied config rule: prepended context for '"+rule.Match+"'")
		}
		if rule.Append != "" {
			text = text + "\n\n" + rule.Append
			improvements = append(improvements, "Applied config rule: appended context for '"+rule.Match+"'")
		}
	}

	return text, improvements
}
