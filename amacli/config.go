package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type configuredUser struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type pendingAuthState struct {
	DeviceCode              string `json:"device_code"`
	CodeVerifier            string `json:"code_verifier"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresAt               string `json:"expires_at"`
	Interval                int    `json:"interval"`
	ClientName              string `json:"client_name"`
	DeviceName              string `json:"device_name,omitempty"`
}

type localConfig struct {
	BaseURL       string            `json:"base_url,omitempty"`
	APIKey        string            `json:"api_key,omitempty"`
	DefaultSource string            `json:"default_source,omitempty"`
	User          *configuredUser   `json:"user,omitempty"`
	PendingAuth   *pendingAuthState `json:"pending_auth,omitempty"`
	UpdatedAt     string            `json:"updated_at,omitempty"`
}

func defaultConfigPath(getenv envGetter) string {
	if getenv != nil {
		if configured := strings.TrimSpace(getenv("AMA_CONFIG_PATH")); configured != "" {
			return configured
		}
	}

	homeDir, err := os.UserHomeDir()
	if err == nil && strings.TrimSpace(homeDir) != "" {
		return filepath.Join(homeDir, ".config", "amacli", "config.json")
	}

	return filepath.Join(".", "amacli-config.json")
}

func readLocalConfig(path string) (localConfig, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return localConfig{}, fmt.Errorf("config path is required")
	}

	body, err := os.ReadFile(trimmed)
	if err != nil {
		if os.IsNotExist(err) {
			return localConfig{}, nil
		}

		return localConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg localConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return localConfig{}, fmt.Errorf("decode config: %w", err)
	}

	return cfg, nil
}

func writeLocalConfig(path string, cfg localConfig) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return fmt.Errorf("config path is required")
	}

	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	directory := filepath.Dir(trimmed)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	tempPath := trimmed + ".tmp"
	if err := os.WriteFile(tempPath, append(body, '\n'), 0o600); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}

	if err := os.Rename(tempPath, trimmed); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}

	return nil
}
