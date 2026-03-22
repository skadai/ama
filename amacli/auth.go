package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type startAuthRequest struct {
	ClientName          string `json:"client_name"`
	DeviceName          string `json:"device_name,omitempty"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type startAuthResponse struct {
	RequestID               string `json:"request_id"`
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type claimAuthRequest struct {
	DeviceCode   string `json:"device_code"`
	CodeVerifier string `json:"code_verifier"`
}

type claimAuthResponse struct {
	RequestID string          `json:"request_id"`
	Status    string          `json:"status"`
	APIKey    string          `json:"api_key,omitempty"`
	User      *configuredUser `json:"user,omitempty"`
	Access    *struct {
		Sources []string `json:"sources"`
	} `json:"access,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
}

func executeAuth(
	ctx context.Context,
	client *amaClient,
	configPath string,
	cfg localConfig,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	if len(args) == 0 {
		printAuthUsage(stderr)
		return errors.New("missing auth subcommand")
	}

	switch args[0] {
	case "login":
		return executeAuthLogin(ctx, client, configPath, cfg, args[1:], stdout, stderr)
	case "complete":
		return executeAuthComplete(ctx, client, configPath, cfg, stdout)
	case "status":
		return executeAuthStatus(configPath, cfg, stdout)
	case "logout":
		return executeAuthLogout(configPath, cfg, stdout)
	default:
		printAuthUsage(stderr)
		return fmt.Errorf("unknown auth subcommand: %s", args[0])
	}
}

func executeAuthLogin(
	ctx context.Context,
	client *amaClient,
	configPath string,
	cfg localConfig,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	loginFlags := flag.NewFlagSet("auth login", flag.ContinueOnError)
	loginFlags.SetOutput(stderr)

	hostname, _ := os.Hostname()
	var clientName string
	var deviceName string
	var wait bool

	loginFlags.StringVar(&clientName, "client-name", "Ama CLI", "client name shown during browser approval")
	loginFlags.StringVar(&deviceName, "device-name", firstNonEmpty(hostname, "This device"), "device name shown during browser approval")
	loginFlags.BoolVar(&wait, "wait", false, "wait for browser approval and finish login automatically")

	if err := loginFlags.Parse(args); err != nil {
		return err
	}

	codeVerifier, err := generatePKCEVerifier()
	if err != nil {
		return err
	}

	var response startAuthResponse
	if err := client.postPublic(ctx, "/v1/cli/auth/start", startAuthRequest{
		ClientName:          strings.TrimSpace(clientName),
		DeviceName:          strings.TrimSpace(deviceName),
		CodeChallenge:       buildCodeChallenge(codeVerifier),
		CodeChallengeMethod: "S256",
	}, &response); err != nil {
		return err
	}

	cfg.BaseURL = client.baseURL.String()
	cfg.PendingAuth = &pendingAuthState{
		DeviceCode:              response.DeviceCode,
		CodeVerifier:            codeVerifier,
		UserCode:                response.UserCode,
		VerificationURI:         response.VerificationURI,
		VerificationURIComplete: response.VerificationURIComplete,
		ExpiresAt:               time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).UTC().Format(time.RFC3339),
		Interval:                response.Interval,
		ClientName:              strings.TrimSpace(clientName),
		DeviceName:              strings.TrimSpace(deviceName),
	}

	if err := writeLocalConfig(configPath, cfg); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "Open this URL in your browser and approve the CLI session:\n%s\n\nUser code: %s\n", response.VerificationURIComplete, response.UserCode)

	if wait {
		_, _ = fmt.Fprintln(stdout, "\nWaiting for browser approval...")
		return waitForAuthCompletion(ctx, client, configPath, cfg, stdout)
	}

	_, _ = fmt.Fprintln(stdout, "\nAfter approving in the browser, run: amacli auth complete")
	return nil
}

func executeAuthComplete(
	ctx context.Context,
	client *amaClient,
	configPath string,
	cfg localConfig,
	stdout io.Writer,
) error {
	if cfg.PendingAuth == nil {
		return errors.New("no pending browser authorization found; run `amacli auth login` first")
	}

	if pendingExpired(cfg.PendingAuth) {
		cfg.PendingAuth = nil
		if err := writeLocalConfig(configPath, cfg); err != nil {
			return err
		}
		return errors.New("the pending browser authorization has expired; run `amacli auth login` again")
	}

	var response claimAuthResponse
	if err := client.postPublic(ctx, "/v1/cli/auth/claim", claimAuthRequest{
		DeviceCode:   cfg.PendingAuth.DeviceCode,
		CodeVerifier: cfg.PendingAuth.CodeVerifier,
	}, &response); err != nil {
		return err
	}

	switch response.Status {
	case "approved":
		cfg.APIKey = strings.TrimSpace(response.APIKey)
		cfg.BaseURL = firstNonEmpty(response.BaseURL, cfg.BaseURL, client.baseURL.String())
		allowedSources := []string{}
		if response.Access != nil {
			allowedSources = uniqueSources(response.Access.Sources)
		}
		cfg.DefaultSource = resolveDefaultSource(cfg.DefaultSource, allowedSources)
		cfg.PendingAuth = nil
		if response.User != nil {
			cfg.User = response.User
		}
		if err := writeLocalConfig(configPath, cfg); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(stdout, "Saved API key to %s\n", configPath)
		if response.User != nil {
			_, _ = fmt.Fprintf(stdout, "Authenticated as %s\n", firstNonEmpty(response.User.Email, response.User.Name, response.User.ID))
		}
		return nil
	case "authorization_pending":
		return errors.New("browser approval is still pending; ask the user to finish authorization, then run `amacli auth complete` again")
	case "access_denied":
		cfg.PendingAuth = nil
		_ = writeLocalConfig(configPath, cfg)
		return errors.New("browser authorization was denied; run `amacli auth login` to start over")
	case "expired_token":
		cfg.PendingAuth = nil
		_ = writeLocalConfig(configPath, cfg)
		return errors.New("browser authorization expired; run `amacli auth login` again")
	default:
		return fmt.Errorf("unexpected auth status: %s", response.Status)
	}
}

func executeAuthStatus(configPath string, cfg localConfig, stdout io.Writer) error {
	status := map[string]any{
		"config_path":        configPath,
		"base_url":           cfg.BaseURL,
		"authenticated":      strings.TrimSpace(cfg.APIKey) != "",
		"api_key_configured": strings.TrimSpace(cfg.APIKey) != "",
		"has_pending_auth":   cfg.PendingAuth != nil,
		"default_source":     firstNonEmpty(cfg.DefaultSource, defaultSource),
		"user":               cfg.User,
	}

	if cfg.PendingAuth != nil {
		status["pending_auth"] = map[string]any{
			"user_code":                 cfg.PendingAuth.UserCode,
			"verification_uri_complete": cfg.PendingAuth.VerificationURIComplete,
			"expires_at":                cfg.PendingAuth.ExpiresAt,
			"client_name":               cfg.PendingAuth.ClientName,
			"device_name":               cfg.PendingAuth.DeviceName,
		}
	}

	return writeJSON(stdout, status)
}

func executeAuthLogout(configPath string, cfg localConfig, stdout io.Writer) error {
	cfg.APIKey = ""
	cfg.User = nil
	cfg.PendingAuth = nil
	if err := writeLocalConfig(configPath, cfg); err != nil {
		return err
	}

	_, err := fmt.Fprintf(stdout, "Cleared local authentication state in %s\n", configPath)
	return err
}

func waitForAuthCompletion(
	ctx context.Context,
	client *amaClient,
	configPath string,
	cfg localConfig,
	stdout io.Writer,
) error {
	if cfg.PendingAuth == nil {
		return errors.New("pending auth state is missing")
	}

	interval := time.Duration(cfg.PendingAuth.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(15 * time.Minute)
	if expiresAt, err := time.Parse(time.RFC3339, cfg.PendingAuth.ExpiresAt); err == nil {
		deadline = expiresAt
	}

	for {
		if time.Now().After(deadline) {
			return errors.New("browser authorization expired before completion")
		}

		err := executeAuthComplete(ctx, client, configPath, cfg, stdout)
		if err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "still pending") {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			refreshedCfg, readErr := readLocalConfig(configPath)
			if readErr == nil {
				cfg = refreshedCfg
			}
		}
	}
}

func pendingExpired(pending *pendingAuthState) bool {
	if pending == nil {
		return false
	}

	expiresAt, err := time.Parse(time.RFC3339, pending.ExpiresAt)
	if err != nil {
		return false
	}

	return time.Now().After(expiresAt)
}

func generatePKCEVerifier() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate PKCE verifier: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func buildCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func printAuthUsage(writer io.Writer) {
	_, _ = fmt.Fprintf(writer, `Auth usage:
  amacli auth login [--wait]
  amacli auth complete
  amacli auth status
  amacli auth logout
`)
}
