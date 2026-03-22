package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHealthCommand(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/health" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"ok","mode":"search-only"}`))
	}))
	defer server.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--base-url", server.URL, "health"},
		stdout,
		stderr,
		func(string) string { return "" },
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"status": "ok"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestSearchCommandUsesDefaultsAndHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/search" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		if got := request.Header.Get("x-ama-client-type"); got != "cli" {
			t.Fatalf("unexpected client type header: %s", got)
		}

		var payload searchRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}

		if payload.Query != "product mvp" {
			t.Fatalf("unexpected query: %+v", payload)
		}
		if payload.TopK != 5 {
			t.Fatalf("unexpected topK: %+v", payload)
		}
		if len(payload.Sources) != 1 || payload.Sources[0] != defaultSource {
			t.Fatalf("unexpected sources: %+v", payload.Sources)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"results":[{"id":42,"title":"Test"}]}`))
	}))
	defer server.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--base-url", server.URL, "search", "--top-k", "5", "product mvp"},
		stdout,
		stderr,
		func(key string) string {
			if key == "AMA_API_KEY" {
				return "test-key"
			}
			return ""
		},
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"id": 42`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestDocumentCommandSupportsPositionalArguments(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/documents/lenny/42" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"document":{"id":42,"title":"Doc"}}`))
	}))
	defer server.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--base-url", server.URL, "doc", "lenny", "42"},
		stdout,
		stderr,
		func(key string) string {
			if key == "AMA_API_KEY" {
				return "test-key"
			}
			return ""
		},
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"title": "Doc"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestAuthLoginWritesPendingConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/cli/auth/start" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		var payload startAuthRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}

		if payload.ClientName != "Ama CLI" {
			t.Fatalf("unexpected client name: %+v", payload)
		}
		if payload.CodeChallengeMethod != "S256" {
			t.Fatalf("unexpected challenge method: %+v", payload)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"request_id":"req_1","device_code":"device-code","user_code":"ABCD-EFGH","verification_uri":"http://localhost:3000/auth/cli/activate","verification_uri_complete":"http://localhost:3000/auth/cli/activate?user_code=ABCD-EFGH","expires_in":900,"interval":5}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "--base-url", server.URL, "auth", "login"},
		stdout,
		stderr,
		func(string) string { return "" },
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	cfg, err := readLocalConfig(configPath)
	if err != nil {
		t.Fatalf("readLocalConfig returned error: %v", err)
	}
	if cfg.PendingAuth == nil {
		t.Fatal("expected pending auth to be written")
	}
	if cfg.PendingAuth.UserCode != "ABCD-EFGH" {
		t.Fatalf("unexpected pending auth: %+v", cfg.PendingAuth)
	}
	if !strings.Contains(stdout.String(), "Open this URL") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestAuthCompleteStoresAPIKey(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/cli/auth/claim" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		var payload claimAuthRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if payload.DeviceCode != "device-code" {
			t.Fatalf("unexpected payload: %+v", payload)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"request_id":"req_2","status":"approved","api_key":"amk_live_secret","user":{"id":"user_123","email":"founder@example.com","name":"Founder"},"access":{"sources":["other-source","lenny"]},"base_url":"http://localhost:3000"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := localConfig{
		BaseURL: server.URL,
		PendingAuth: &pendingAuthState{
			DeviceCode:              "device-code",
			CodeVerifier:            "code-verifier",
			UserCode:                "ABCD-EFGH",
			VerificationURI:         "http://localhost:3000/auth/cli/activate",
			VerificationURIComplete: "http://localhost:3000/auth/cli/activate?user_code=ABCD-EFGH",
			ExpiresAt:               "2099-01-01T00:00:00Z",
			Interval:                5,
			ClientName:              "Ama CLI",
			DeviceName:              "Test Device",
		},
	}
	if err := writeLocalConfig(configPath, cfg); err != nil {
		t.Fatalf("writeLocalConfig returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "--base-url", server.URL, "auth", "complete"},
		stdout,
		stderr,
		func(string) string { return "" },
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	updatedCfg, err := readLocalConfig(configPath)
	if err != nil {
		t.Fatalf("readLocalConfig returned error: %v", err)
	}
	if updatedCfg.APIKey != "amk_live_secret" {
		t.Fatalf("unexpected api key: %+v", updatedCfg)
	}
	if updatedCfg.PendingAuth != nil {
		t.Fatalf("expected pending auth to be cleared: %+v", updatedCfg.PendingAuth)
	}
	if updatedCfg.User == nil || updatedCfg.User.Email != "founder@example.com" {
		t.Fatalf("unexpected user: %+v", updatedCfg.User)
	}
	if updatedCfg.DefaultSource != "other-source" {
		t.Fatalf("unexpected default source: %+v", updatedCfg)
	}
	if !strings.Contains(stdout.String(), "Saved API key") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestProtectedCommandsRequireAPIKey(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	err := run(
		context.Background(),
		[]string{"--config", configPath, "search", "mvp"},
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(string) string { return "" },
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "AMA_API_KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteLocalConfigCreatesA0600File(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := writeLocalConfig(configPath, localConfig{BaseURL: defaultBaseURL}); err != nil {
		t.Fatalf("writeLocalConfig returned error: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat returned error: %v", err)
	}

	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected permissions: %v", info.Mode().Perm())
	}
}

func TestDefaultConfigPathUsesDotConfigAmacli(t *testing.T) {
	t.Parallel()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir returned error: %v", err)
	}

	got := defaultConfigPath(func(string) string { return "" })
	want := filepath.Join(homeDir, ".config", "amacli", "config.json")
	if got != want {
		t.Fatalf("unexpected default config path: got %s want %s", got, want)
	}
}

func TestSaveAnswerCommandPostsStructuredPayload(t *testing.T) {
	t.Parallel()

	answerPath := filepath.Join(t.TempDir(), "answer.md")
	if err := os.WriteFile(answerPath, []byte("Final answer with citations."), 0o600); err != nil {
		t.Fatalf("WriteFile answer: %v", err)
	}

	citationsPath := filepath.Join(t.TempDir(), "citations.json")
	if err := os.WriteFile(citationsPath, []byte(`[{"title":"Great PM hiring","type":"podcast_episode","date":"2025-02-01","source_slug":"lenny"}]`), 0o600); err != nil {
		t.Fatalf("WriteFile citations: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/saved-answers" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}

		var payload saveAnswerRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}

		if payload.Question != "What does Lenny say about PM hiring?" {
			t.Fatalf("unexpected question: %+v", payload)
		}
		if payload.Answer != "Final answer with citations." {
			t.Fatalf("unexpected answer: %+v", payload)
		}
		if len(payload.Citations) != 1 || payload.Citations[0]["title"] != "Great PM hiring" {
			t.Fatalf("unexpected citations: %+v", payload.Citations)
		}
		if len(payload.SourceSlugs) != 1 || payload.SourceSlugs[0] != defaultSource {
			t.Fatalf("unexpected sources: %+v", payload.SourceSlugs)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"request_id":"req_saved","saved_answer":{"id":"save_123"}}`))
	}))
	defer server.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--base-url", server.URL, "save-answer", "--question", "What does Lenny say about PM hiring?", "--answer-file", answerPath, "--citations-file", citationsPath},
		stdout,
		stderr,
		func(key string) string {
			if key == "AMA_API_KEY" {
				return "test-key"
			}
			return ""
		},
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"id": "save_123"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestSaveAnswerCommandRequiresAnswerInput(t *testing.T) {
	t.Parallel()

	err := run(
		context.Background(),
		[]string{"save-answer", "--question", "What does Lenny say about PM hiring?"},
		&bytes.Buffer{},
		&bytes.Buffer{},
		func(key string) string {
			if key == "AMA_API_KEY" {
				return "test-key"
			}
			return ""
		},
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "answer is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentCommandUsesConfiguredDefaultSource(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/documents/custom-source/42" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"document":{"id":42,"title":"Doc"}}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := writeLocalConfig(configPath, localConfig{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		DefaultSource: "custom-source",
	}); err != nil {
		t.Fatalf("writeLocalConfig returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "doc", "42"},
		stdout,
		stderr,
		func(string) string { return "" },
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"title": "Doc"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestSourceListUsesMeEndpoint(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/me" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"request_id":"req_me","access":{"sources":["lenny","other-source","lenny"]}}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := writeLocalConfig(configPath, localConfig{
		BaseURL:       server.URL,
		APIKey:        "test-key",
		DefaultSource: "other-source",
	}); err != nil {
		t.Fatalf("writeLocalConfig returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "source", "list"},
		stdout,
		stderr,
		func(string) string { return "" },
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"default_source": "other-source"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"sources": [`) || strings.Count(stdout.String(), `"lenny"`) != 1 {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestSourceSetDefaultValidatesAndWritesConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.URL.Path != "/v1/me" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"request_id":"req_me","access":{"sources":["lenny","other-source"]}}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := writeLocalConfig(configPath, localConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	}); err != nil {
		t.Fatalf("writeLocalConfig returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "source", "set-default", "other-source"},
		stdout,
		stderr,
		func(string) string { return "" },
		server.Client(),
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	cfg, err := readLocalConfig(configPath)
	if err != nil {
		t.Fatalf("readLocalConfig returned error: %v", err)
	}
	if cfg.DefaultSource != "other-source" {
		t.Fatalf("unexpected default source: %+v", cfg)
	}
	if !strings.Contains(stdout.String(), `Saved default source "other-source"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestLanguageSetWritesConfig(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "language", "set", "zh"},
		stdout,
		stderr,
		func(string) string { return "" },
		nil,
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	cfg, err := readLocalConfig(configPath)
	if err != nil {
		t.Fatalf("readLocalConfig returned error: %v", err)
	}
	if cfg.PreferredLanguage != "zh" {
		t.Fatalf("unexpected preferred language: %+v", cfg)
	}
	if !strings.Contains(stdout.String(), `Saved preferred language "zh"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestLanguageShowReturnsStoredValue(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := writeLocalConfig(configPath, localConfig{PreferredLanguage: "en"}); err != nil {
		t.Fatalf("writeLocalConfig returned error: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := run(
		context.Background(),
		[]string{"--config", configPath, "language", "show"},
		stdout,
		stderr,
		func(string) string { return "" },
		nil,
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), `"preferred_language": "en"`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestNewAmaClientUsesDefaultTimeout(t *testing.T) {
	t.Parallel()

	client, err := newAmaClient(config{BaseURL: defaultBaseURL}, nil)
	if err != nil {
		t.Fatalf("newAmaClient returned error: %v", err)
	}

	if client.httpClient.Timeout != defaultHTTPTimeout {
		t.Fatalf("unexpected timeout: got %s want %s", client.httpClient.Timeout, defaultHTTPTimeout)
	}
}

func TestResolveHTTPTimeoutUsesEnvOverride(t *testing.T) {
	t.Parallel()

	timeout, err := resolveHTTPTimeout(0, "90s")
	if err != nil {
		t.Fatalf("resolveHTTPTimeout returned error: %v", err)
	}

	if timeout != 90*time.Second {
		t.Fatalf("unexpected timeout: got %s want %s", timeout, 90*time.Second)
	}
}
