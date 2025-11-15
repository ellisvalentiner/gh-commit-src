package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	OpenAI struct {
		APIKey     string `yaml:"api_key" env:"OPENAI_API_KEY"`
		URL        string `yaml:"url" env:"OPENAI_URL"`
		Model      string `yaml:"model" env:"OPENAI_MODEL"`
		APIVersion string `yaml:"api_version" env:"AZURE_API_VERSION"`
	} `yaml:"openai"`

	FineTune struct {
		Temperature      *float32 `yaml:"temperature"`
		MaxTokens        *int32   `yaml:"max_tokens"`
		TopP             *float32 `yaml:"top_p"`
		FrequencyPenalty *float32 `yaml:"frequency_penalty"`
		PresencePenalty  *float32 `yaml:"presence_penalty"`
	} `yaml:"fine_tune"`

	Prompt struct {
		Override string `yaml:"override" env:"PROMPT_OVERRIDE"`
		Suffix   string `yaml:"suffix" env:"COMMIT_MESSAGE_SUFFIX"`
	} `yaml:"prompt"`

	CodeBlock struct {
		Patterns []string `yaml:"patterns" env:"CODE_BLOCK_PATTERNS"`
	} `yaml:"code_block"`
}

var globalConfig *Config

// resetConfig resets the global config (useful for testing)
func resetConfig() {
	globalConfig = nil
}

// loadConfig loads configuration from file and environment variables
// Environment variables take precedence over config file values
func loadConfig() (*Config, error) {
	config := &Config{}

	// Set defaults
	config.OpenAI.URL = "https://api.openai.com/v1"
	config.OpenAI.Model = "gpt-4"
	config.OpenAI.APIVersion = defaultAzureAPIVersion
	config.Prompt.Suffix = defaultCommitMessageSuffix
	config.CodeBlock.Patterns = defaultCodeBlockPatterns

	// Try to load from config file
	configPath := findConfigFile()
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, fmt.Errorf("error parsing config file %s: %v", configPath, err)
			}
		}
	}

	// Override with environment variables (env vars take precedence)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.OpenAI.APIKey = apiKey
	}
	if url := os.Getenv("OPENAI_URL"); url != "" {
		config.OpenAI.URL = url
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		config.OpenAI.Model = model
	}
	if version := os.Getenv("AZURE_API_VERSION"); version != "" {
		config.OpenAI.APIVersion = version
	}
	if prompt := os.Getenv("PROMPT_OVERRIDE"); prompt != "" {
		config.Prompt.Override = prompt
	}
	if suffix := os.Getenv("COMMIT_MESSAGE_SUFFIX"); suffix != "" {
		config.Prompt.Suffix = suffix
	}
	if patterns := os.Getenv("CODE_BLOCK_PATTERNS"); patterns != "" {
		// Parse comma-separated patterns
		patternList := []string{}
		for _, p := range splitAndTrim(patterns, ",") {
			if p != "" {
				patternList = append(patternList, p)
			}
		}
		if len(patternList) > 0 {
			config.CodeBlock.Patterns = patternList
		}
	}

	// Parse FINE_TUNE_PARAMS from environment if present
	if fineTuneParams := os.Getenv("FINE_TUNE_PARAMS"); fineTuneParams != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(fineTuneParams), &params); err == nil {
			if temp, ok := params["temperature"].(float64); ok {
				t := float32(temp)
				config.FineTune.Temperature = &t
			}
			if maxTokens, ok := params["max_tokens"].(float64); ok {
				mt := int32(maxTokens)
				config.FineTune.MaxTokens = &mt
			}
			if topP, ok := params["top_p"].(float64); ok {
				tp := float32(topP)
				config.FineTune.TopP = &tp
			}
			if freqPenalty, ok := params["frequency_penalty"].(float64); ok {
				fp := float32(freqPenalty)
				config.FineTune.FrequencyPenalty = &fp
			}
			if presPenalty, ok := params["presence_penalty"].(float64); ok {
				pp := float32(presPenalty)
				config.FineTune.PresencePenalty = &pp
			}
		}
	}

	// Fine-tune params from config file are already loaded above when parsing the YAML
	// Environment variables (FINE_TUNE_PARAMS) take precedence and are applied above

	return config, nil
}

// findConfigFile searches for config file in standard locations
func findConfigFile() string {
	// 1. Current directory
	configFiles := []string{
		".gh-commit-src.yaml",
		".gh-commit-src.yml",
		"gh-commit-src.yaml",
		"gh-commit-src.yml",
	}

	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	// 2. Home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		for _, filename := range configFiles {
			path := filepath.Join(homeDir, filename)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// 3. XDG config directory
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configDir := filepath.Join(xdgConfig, "gh-commit-src")
		for _, filename := range []string{"config.yaml", "config.yml"} {
			path := filepath.Join(configDir, filename)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	} else {
		// Fallback to ~/.config
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(homeDir, ".config", "gh-commit-src")
			for _, filename := range []string{"config.yaml", "config.yml"} {
				path := filepath.Join(configDir, filename)
				if _, err := os.Stat(path); err == nil {
					return path
				}
			}
		}
	}

	return ""
}

// getConfig returns the global config, loading it if necessary
func getConfig() (*Config, error) {
	if globalConfig == nil {
		var err error
		globalConfig, err = loadConfig()
		if err != nil {
			return nil, err
		}
	}
	return globalConfig, nil
}

// splitAndTrim splits a string by separator and trims whitespace from each part
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}
