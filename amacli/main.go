package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL     = "https://askmeanything.pro"
	defaultSource      = "lenny"
	defaultHTTPTimeout = 60 * time.Second
)

var version = "dev"

type envGetter func(string) string

type config struct {
	BaseURL     string
	APIKey      string
	HTTPTimeout time.Duration
}

type stringList []string

func (items *stringList) String() string {
	return strings.Join(*items, ",")
}

func (items *stringList) Set(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New("value cannot be empty")
	}

	*items = append(*items, trimmed)
	return nil
}

type searchRequest struct {
	Query        string   `json:"query"`
	Sources      []string `json:"sources,omitempty"`
	TopK         int      `json:"top_k,omitempty"`
	ContentTypes []string `json:"content_types,omitempty"`
}

type saveAnswerRequest struct {
	Question    string           `json:"question"`
	Answer      string           `json:"answer"`
	Citations   []map[string]any `json:"citations,omitempty"`
	SourceSlugs []string         `json:"source_slugs,omitempty"`
}

type apiErrorResponse struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type meResponse struct {
	RequestID string `json:"request_id,omitempty"`
	Access    struct {
		Sources []string `json:"sources"`
	} `json:"access"`
}

type amaClient struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

func newAmaClient(cfg config, httpClient *http.Client) (*amaClient, error) {
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		base = defaultBaseURL
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid base url: %s", base)
	}

	timeout := cfg.HTTPTimeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &amaClient{
		baseURL:    parsed,
		apiKey:     strings.TrimSpace(cfg.APIKey),
		httpClient: httpClient,
	}, nil
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, os.Getenv, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	getenv envGetter,
	httpClient *http.Client,
) error {
	if getenv == nil {
		return errors.New("getenv is required")
	}

	global := flag.NewFlagSet("amacli", flag.ContinueOnError)
	global.SetOutput(stderr)
	global.Usage = func() {
		printUsage(stderr)
	}

	var configPath string
	var baseURLFlag string
	var apiKeyFlag string
	var timeoutFlag time.Duration

	configPath = defaultConfigPath(getenv)
	global.StringVar(&configPath, "config", configPath, "path to amacli config")
	global.StringVar(&baseURLFlag, "base-url", "", "AMA API base URL")
	global.StringVar(&apiKeyFlag, "api-key", "", "AMA API key")
	global.DurationVar(&timeoutFlag, "timeout", 0, "HTTP timeout (for example 30s, 1m, 90s)")

	if err := global.Parse(args); err != nil {
		return err
	}

	rest := global.Args()
	if len(rest) == 0 {
		printUsage(stderr)
		return errors.New("missing command")
	}

	if rest[0] == "help" {
		printUsage(stdout)
		return nil
	}

	if rest[0] == "version" {
		_, err := fmt.Fprintln(stdout, version)
		return err
	}

	localCfg, err := readLocalConfig(configPath)
	if err != nil {
		return err
	}

	baseURL := firstNonEmpty(baseURLFlag, getenv("AMA_BASE_URL"), localCfg.BaseURL, defaultBaseURL)
	apiKey := firstNonEmpty(apiKeyFlag, getenv("AMA_API_KEY"), localCfg.APIKey)
	httpTimeout, err := resolveHTTPTimeout(timeoutFlag, getenv("AMA_HTTP_TIMEOUT"))
	if err != nil {
		return err
	}

	client, err := newAmaClient(config{BaseURL: baseURL, APIKey: apiKey, HTTPTimeout: httpTimeout}, httpClient)
	if err != nil {
		return err
	}

	switch rest[0] {
	case "health":
		return executeHealth(ctx, client, stdout)
	case "me":
		if err := requireAPIKey(client.apiKey); err != nil {
			return err
		}

		return executeMe(ctx, client, stdout)
	case "search":
		if err := requireAPIKey(client.apiKey); err != nil {
			return err
		}

		return executeSearch(ctx, client, rest[1:], stdout, stderr)
	case "document", "doc":
		if err := requireAPIKey(client.apiKey); err != nil {
			return err
		}

		return executeDocument(ctx, client, rest[1:], stdout, stderr, localCfg)
	case "save-answer", "save":
		if err := requireAPIKey(client.apiKey); err != nil {
			return err
		}

		return executeSaveAnswer(ctx, client, rest[1:], stdout, stderr, localCfg)
	case "source", "sources":
		if err := requireAPIKey(client.apiKey); err != nil {
			return err
		}

		return executeSource(ctx, client, configPath, localCfg, rest[1:], stdout, stderr)
	case "language", "lang":
		return executeLanguage(configPath, localCfg, rest[1:], stdout, stderr)
	case "auth":
		return executeAuth(ctx, client, configPath, localCfg, rest[1:], stdout, stderr)
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command: %s", rest[0])
	}
}

func executeHealth(ctx context.Context, client *amaClient, stdout io.Writer) error {
	var response any
	if err := client.get(ctx, "/v1/health", true, &response); err != nil {
		return err
	}

	return writeJSON(stdout, response)
}

func executeMe(ctx context.Context, client *amaClient, stdout io.Writer) error {
	var response any
	if err := client.get(ctx, "/v1/me", false, &response); err != nil {
		return err
	}

	return writeJSON(stdout, response)
}

type searchAPIResponse struct {
	RequestID string           `json:"request_id,omitempty"`
	Query     string           `json:"query,omitempty"`
	Terms     []string         `json:"terms,omitempty"`
	Keywords  []string         `json:"keywords,omitempty"`
	Results   []map[string]any `json:"results,omitempty"`
	Usage     map[string]any   `json:"usage,omitempty"`
}

func executeSearch(
	ctx context.Context,
	client *amaClient,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	searchFlags := flag.NewFlagSet("search", flag.ContinueOnError)
	searchFlags.SetOutput(stderr)

	var query string
	var topK int
	var sources stringList
	var contentTypes stringList
	var balancedContentTypes bool

	searchFlags.StringVar(&query, "query", "", "search query")
	searchFlags.StringVar(&query, "q", "", "search query")
	searchFlags.IntVar(&topK, "top-k", 8, "max result count")
	searchFlags.Var(&sources, "source", "source slug filter (repeatable); omit or use \"all\" to search every source")
	searchFlags.Var(&contentTypes, "content-type", "content type filter (repeatable)")
	searchFlags.BoolVar(&balancedContentTypes, "balanced-content-types", false, "run one search per content type and merge results to keep newsletters and podcasts represented")

	if err := searchFlags.Parse(args); err != nil {
		return err
	}

	if query == "" && len(searchFlags.Args()) > 0 {
		query = strings.Join(searchFlags.Args(), " ")
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("search query is required")
	}

	request := searchRequest{
		Query: query,
		TopK:  topK,
	}

	if resolvedSources := resolveSearchSources([]string(sources)); len(resolvedSources) > 0 {
		request.Sources = resolvedSources
	}

	if balancedContentTypes {
		return executeBalancedSearch(ctx, client, request, []string(contentTypes), stdout)
	}

	if len(contentTypes) > 0 {
		request.ContentTypes = []string(contentTypes)
	}

	var response any
	if err := client.post(ctx, "/v1/search", request, &response); err != nil {
		return err
	}

	return writeJSON(stdout, response)
}

func executeBalancedSearch(
	ctx context.Context,
	client *amaClient,
	request searchRequest,
	contentTypes []string,
	stdout io.Writer,
) error {
	balancedTypes := uniqueStringsPreserveOrder(contentTypes)
	if len(balancedTypes) == 0 {
		balancedTypes = []string{"newsletter_article", "podcast_episode"}
	}

	groups := make(map[string][]map[string]any, len(balancedTypes))
	requestSummaries := make([]map[string]any, 0, len(balancedTypes))
	terms := []string{}
	keywords := []string{}

	for _, contentType := range balancedTypes {
		typedRequest := request
		typedRequest.ContentTypes = []string{contentType}

		var response searchAPIResponse
		if err := client.post(ctx, "/v1/search", typedRequest, &response); err != nil {
			return err
		}

		groups[contentType] = response.Results
		terms = append(terms, response.Terms...)
		keywords = append(keywords, response.Keywords...)
		requestSummaries = append(requestSummaries, map[string]any{
			"content_type": contentType,
			"request_id":   response.RequestID,
			"result_count": len(response.Results),
		})
	}

	mergedResults := interleaveSearchResults(groups, balancedTypes, request.TopK)

	response := map[string]any{
		"query":                  request.Query,
		"strategy":               "balanced_content_types",
		"balanced_content_types": balancedTypes,
		"results":                mergedResults,
		"requests":               requestSummaries,
		"usage": map[string]any{
			"result_count":             len(mergedResults),
			"per_type_requested_top_k": request.TopK,
		},
	}
	if len(request.Sources) > 0 {
		response["sources"] = request.Sources
	}

	if dedupedTerms := uniqueStringsPreserveOrder(terms); len(dedupedTerms) > 0 {
		response["terms"] = dedupedTerms
	}
	if dedupedKeywords := uniqueStringsPreserveOrder(keywords); len(dedupedKeywords) > 0 {
		response["keywords"] = dedupedKeywords
	}

	return writeJSON(stdout, response)
}

func interleaveSearchResults(
	groups map[string][]map[string]any,
	order []string,
	limit int,
) []map[string]any {
	if limit <= 0 {
		return nil
	}

	positions := make(map[string]int, len(order))
	seen := make(map[string]struct{})
	results := make([]map[string]any, 0, limit)

	for len(results) < limit {
		progress := false
		for _, contentType := range order {
			items := groups[contentType]
			for positions[contentType] < len(items) {
				candidate := items[positions[contentType]]
				positions[contentType]++
				key := searchResultKey(candidate)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				results = append(results, candidate)
				progress = true
				break
			}
			if len(results) >= limit {
				break
			}
		}
		if !progress {
			break
		}
	}

	return results
}

func searchResultKey(result map[string]any) string {
	idPart := strings.TrimSpace(fmt.Sprint(result["id"]))
	sourcePart := strings.TrimSpace(fmt.Sprint(result["source_slug"]))
	if sourcePart != "" || idPart != "" {
		return sourcePart + ":" + idPart
	}
	return strings.TrimSpace(fmt.Sprint(result["title"]))
}

func uniqueStringsPreserveOrder(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func resolveSearchSources(values []string) []string {
	normalized := uniqueStringsPreserveOrder(values)
	if len(normalized) == 0 {
		return nil
	}

	for _, value := range normalized {
		if strings.EqualFold(strings.TrimSpace(value), "all") {
			return nil
		}
	}

	return normalized
}

func executeDocument(
	ctx context.Context,
	client *amaClient,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	localCfg localConfig,
) error {
	documentFlags := flag.NewFlagSet("document", flag.ContinueOnError)
	documentFlags.SetOutput(stderr)

	var source string
	var articleID int

	documentFlags.StringVar(&source, "source", "", "source slug")
	documentFlags.IntVar(&articleID, "id", 0, "article id")

	if err := documentFlags.Parse(args); err != nil {
		return err
	}

	positionals := documentFlags.Args()
	if strings.TrimSpace(source) == "" {
		switch len(positionals) {
		case 0:
		case 1:
			if parsedID, err := strconv.Atoi(positionals[0]); err == nil {
				articleID = parsedID
			} else {
				source = positionals[0]
			}
		default:
			source = positionals[0]
			if articleID == 0 {
				parsedID, err := strconv.Atoi(positionals[1])
				if err != nil {
					return fmt.Errorf("invalid article id: %s", positionals[1])
				}
				articleID = parsedID
			}
		}
	} else if articleID == 0 && len(positionals) >= 1 {
		parsedID, err := strconv.Atoi(positionals[0])
		if err != nil {
			return fmt.Errorf("invalid article id: %s", positionals[0])
		}
		articleID = parsedID
	}

	source = strings.TrimSpace(source)
	if source == "" {
		source = firstNonEmpty(localCfg.DefaultSource, defaultSource)
	}
	if articleID < 1 {
		return errors.New("document id must be a positive integer")
	}

	endpoint := fmt.Sprintf("/v1/documents/%s/%d", url.PathEscape(source), articleID)
	var response any
	if err := client.get(ctx, endpoint, false, &response); err != nil {
		return err
	}

	return writeJSON(stdout, response)
}

func executeSource(
	ctx context.Context,
	client *amaClient,
	configPath string,
	cfg localConfig,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	if len(args) == 0 {
		printSourceUsage(stderr)
		return errors.New("missing source subcommand")
	}

	switch args[0] {
	case "list", "ls":
		return executeSourceList(ctx, client, cfg, stdout)
	case "set-default", "set":
		return executeSourceSetDefault(ctx, client, configPath, cfg, args[1:], stdout, stderr)
	default:
		printSourceUsage(stderr)
		return fmt.Errorf("unknown source subcommand: %s", args[0])
	}
}

func executeSourceList(
	ctx context.Context,
	client *amaClient,
	cfg localConfig,
	stdout io.Writer,
) error {
	me, err := fetchMe(ctx, client)
	if err != nil {
		return err
	}

	return writeJSON(stdout, map[string]any{
		"sources":        me.Access.Sources,
		"default_source": resolveDefaultSource(cfg.DefaultSource, me.Access.Sources),
	})
}

func executeLanguage(
	configPath string,
	cfg localConfig,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	if len(args) == 0 {
		printLanguageUsage(stderr)
		return errors.New("missing language subcommand")
	}

	switch args[0] {
	case "show", "get", "current":
		return executeLanguageShow(cfg, stdout)
	case "set":
		return executeLanguageSet(configPath, cfg, args[1:], stdout, stderr)
	default:
		printLanguageUsage(stderr)
		return fmt.Errorf("unknown language subcommand: %s", args[0])
	}
}

func executeLanguageShow(cfg localConfig, stdout io.Writer) error {
	value := strings.TrimSpace(cfg.PreferredLanguage)
	if value == "" {
		return writeJSON(stdout, map[string]any{
			"preferred_language": nil,
		})
	}

	return writeJSON(stdout, map[string]any{
		"preferred_language": value,
	})
}

func executeLanguageSet(
	configPath string,
	cfg localConfig,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	setFlags := flag.NewFlagSet("language set", flag.ContinueOnError)
	setFlags.SetOutput(stderr)

	var value string
	setFlags.StringVar(&value, "language", "", "preferred answer language: zh or en")
	setFlags.StringVar(&value, "lang", "", "preferred answer language: zh or en")

	if err := setFlags.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(value) == "" && len(setFlags.Args()) > 0 {
		value = setFlags.Args()[0]
	}

	value = normalizePreferredLanguage(value)
	if value == "" {
		return errors.New("language must be zh or en")
	}

	cfg.PreferredLanguage = value
	if err := writeLocalConfig(configPath, cfg); err != nil {
		return err
	}

	_, err := fmt.Fprintf(stdout, "Saved preferred language %q to %s\n", value, configPath)
	return err
}

func normalizePreferredLanguage(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "zh", "zh-cn", "zh-hans", "cn", "chinese":
		return "zh"
	case "en", "en-us", "en-gb", "english":
		return "en"
	default:
		return ""
	}
}

func executeSourceSetDefault(
	ctx context.Context,
	client *amaClient,
	configPath string,
	cfg localConfig,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	setFlags := flag.NewFlagSet("source set-default", flag.ContinueOnError)
	setFlags.SetOutput(stderr)

	var source string
	setFlags.StringVar(&source, "source", "", "source slug to save as default")

	if err := setFlags.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(source) == "" && len(setFlags.Args()) > 0 {
		source = setFlags.Args()[0]
	}

	source = strings.TrimSpace(source)
	if source == "" {
		return errors.New("source slug is required")
	}

	me, err := fetchMe(ctx, client)
	if err != nil {
		return err
	}

	if !containsSource(me.Access.Sources, source) {
		return fmt.Errorf("source %q is not available for this API key; allowed sources: %s", source, strings.Join(me.Access.Sources, ", "))
	}

	cfg.BaseURL = client.baseURL.String()
	cfg.DefaultSource = source
	if err := writeLocalConfig(configPath, cfg); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "Saved default source %q to %s\n", source, configPath)
	return err
}

func fetchMe(ctx context.Context, client *amaClient) (meResponse, error) {
	var response meResponse
	if err := client.get(ctx, "/v1/me", false, &response); err != nil {
		return meResponse{}, err
	}

	response.Access.Sources = uniqueSources(response.Access.Sources)
	return response, nil
}

func uniqueSources(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func containsSource(values []string, want string) bool {
	trimmedWant := strings.TrimSpace(want)
	for _, value := range values {
		if strings.TrimSpace(value) == trimmedWant {
			return true
		}
	}
	return false
}

func resolveDefaultSource(preferred string, allowed []string) string {
	trimmedPreferred := strings.TrimSpace(preferred)
	if trimmedPreferred != "" && containsSource(allowed, trimmedPreferred) {
		return trimmedPreferred
	}

	if normalized := uniqueSources(allowed); len(normalized) > 0 {
		return normalized[0]
	}

	return firstNonEmpty(trimmedPreferred, defaultSource)
}

func executeSaveAnswer(
	ctx context.Context,
	client *amaClient,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	localCfg localConfig,
) error {
	saveFlags := flag.NewFlagSet("save-answer", flag.ContinueOnError)
	saveFlags.SetOutput(stderr)

	var question string
	var answer string
	var answerFile string
	var citationsJSON string
	var citationsFile string
	var sources stringList

	saveFlags.StringVar(&question, "question", "", "original user question")
	saveFlags.StringVar(&question, "q", "", "original user question")
	saveFlags.StringVar(&answer, "answer", "", "final answer text")
	saveFlags.StringVar(&answerFile, "answer-file", "", "path to a file containing the final answer")
	saveFlags.StringVar(&citationsJSON, "citations-json", "", "inline citations JSON array")
	saveFlags.StringVar(&citationsFile, "citations-file", "", "path to a JSON file containing citations")
	saveFlags.Var(&sources, "source", "source slug (repeatable)")

	if err := saveFlags.Parse(args); err != nil {
		return err
	}

	question = strings.TrimSpace(question)
	if question == "" {
		return errors.New("question is required")
	}

	resolvedAnswer, err := resolveAnswerInput(answer, answerFile)
	if err != nil {
		return err
	}

	citations, err := resolveCitationsInput(citationsJSON, citationsFile)
	if err != nil {
		return err
	}

	request := saveAnswerRequest{
		Question:  question,
		Answer:    resolvedAnswer,
		Citations: citations,
	}

	if len(sources) > 0 {
		request.SourceSlugs = []string(sources)
	} else {
		request.SourceSlugs = []string{firstNonEmpty(localCfg.DefaultSource, defaultSource)}
	}

	var response any
	if err := client.post(ctx, "/v1/saved-answers", request, &response); err != nil {
		return err
	}

	return writeJSON(stdout, response)
}

func resolveAnswerInput(answer string, answerFile string) (string, error) {
	if strings.TrimSpace(answer) != "" && strings.TrimSpace(answerFile) != "" {
		return "", errors.New("use either --answer or --answer-file, not both")
	}

	if strings.TrimSpace(answerFile) != "" {
		body, err := os.ReadFile(strings.TrimSpace(answerFile))
		if err != nil {
			return "", fmt.Errorf("read answer file: %w", err)
		}

		resolved := strings.TrimSpace(string(body))
		if resolved == "" {
			return "", errors.New("answer file is empty")
		}

		return resolved, nil
	}

	if strings.TrimSpace(answer) != "" {
		return strings.TrimSpace(answer), nil
	}

	stdinBody, err := readStdinIfPiped()
	if err != nil {
		return "", err
	}
	if stdinBody != "" {
		return stdinBody, nil
	}

	return "", errors.New("answer is required; pass --answer, --answer-file, or pipe the answer through stdin")
}

func resolveCitationsInput(citationsJSON string, citationsFile string) ([]map[string]any, error) {
	if strings.TrimSpace(citationsJSON) != "" && strings.TrimSpace(citationsFile) != "" {
		return nil, errors.New("use either --citations-json or --citations-file, not both")
	}

	var raw []byte
	if strings.TrimSpace(citationsFile) != "" {
		body, err := os.ReadFile(strings.TrimSpace(citationsFile))
		if err != nil {
			return nil, fmt.Errorf("read citations file: %w", err)
		}
		raw = body
	} else if strings.TrimSpace(citationsJSON) != "" {
		raw = []byte(strings.TrimSpace(citationsJSON))
	} else {
		return nil, nil
	}

	var citations []map[string]any
	if err := json.Unmarshal(raw, &citations); err != nil {
		return nil, fmt.Errorf("decode citations json: %w", err)
	}

	return citations, nil
}

func readStdinIfPiped() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("inspect stdin: %w", err)
	}

	if info.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}

	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

func (client *amaClient) get(ctx context.Context, endpoint string, public bool, out any) error {
	req, err := client.newRequest(ctx, http.MethodGet, endpoint, nil, public)
	if err != nil {
		return err
	}

	return client.do(req, out)
}

func (client *amaClient) post(ctx context.Context, endpoint string, body any, out any) error {
	return client.postWithVisibility(ctx, endpoint, body, out, false)
}

func (client *amaClient) postPublic(ctx context.Context, endpoint string, body any, out any) error {
	return client.postWithVisibility(ctx, endpoint, body, out, true)
}

func (client *amaClient) postWithVisibility(
	ctx context.Context,
	endpoint string,
	body any,
	out any,
	public bool,
) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode request body: %w", err)
	}

	req, err := client.newRequest(ctx, http.MethodPost, endpoint, payload, public)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	return client.do(req, out)
}

func (client *amaClient) newRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body []byte,
	public bool,
) (*http.Request, error) {
	target := *client.baseURL
	target.Path = path.Join(client.baseURL.Path, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "amacli/"+version)
	req.Header.Set("x-ama-client-type", "cli")

	if !public {
		req.Header.Set("Authorization", "Bearer "+client.apiKey)
	}

	return req, nil
}

func (client *amaClient) do(req *http.Request, out any) error {
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return formatAPIError(resp.StatusCode, body)
	}

	if len(body) == 0 {
		return errors.New("empty response body")
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}

func formatAPIError(statusCode int, body []byte) error {
	var parsed apiErrorResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != nil {
		message := strings.TrimSpace(parsed.Error.Message)
		code := strings.TrimSpace(parsed.Error.Code)
		if code != "" {
			return fmt.Errorf("api request failed (%d %s): %s", statusCode, code, message)
		}

		return fmt.Errorf("api request failed (%d): %s", statusCode, message)
	}

	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return fmt.Errorf("api request failed with status %d", statusCode)
	}

	return fmt.Errorf("api request failed with status %d: %s", statusCode, raw)
}

func requireAPIKey(apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("AMA_API_KEY is required for this command")
	}

	return nil
}

func resolveHTTPTimeout(flagValue time.Duration, envValue string) (time.Duration, error) {
	if flagValue > 0 {
		return flagValue, nil
	}

	trimmedEnvValue := strings.TrimSpace(envValue)
	if trimmedEnvValue == "" {
		return defaultHTTPTimeout, nil
	}

	parsed, err := time.ParseDuration(trimmedEnvValue)
	if err != nil {
		return 0, fmt.Errorf("invalid AMA_HTTP_TIMEOUT %q: %w", trimmedEnvValue, err)
	}
	if parsed <= 0 {
		return 0, errors.New("AMA_HTTP_TIMEOUT must be greater than zero")
	}

	return parsed, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func writeJSON(writer io.Writer, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}

	_, err = fmt.Fprintln(writer, string(encoded))
	return err
}

func printSourceUsage(writer io.Writer) {
	_, _ = fmt.Fprintf(writer, `Manage available AMA sources.

Usage:
  amacli source list
  amacli source set-default <source-slug>
  amacli source set-default --source <source-slug>
`)
}

func printLanguageUsage(writer io.Writer) {
	_, _ = fmt.Fprintf(writer, `Manage the preferred answer language.

Usage:
  amacli language show
  amacli language set <zh|en>
  amacli language set --language <zh|en>
`)
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintf(writer, `amacli is a small AMA API client.

Usage:
  amacli [--config PATH] [--base-url URL] [--api-key KEY] <command> [options]

Commands:
  health               Check server health
  me                   Inspect the current API key
  search               Search indexed AMA content
  document|doc         Fetch the original markdown for an article
  save-answer|save     Save a question, final answer, and citations
  source|sources       List allowed sources or set the default source
  language|lang        Show or set the preferred answer language
  auth                 Start or complete browser-based CLI login
  version              Print the CLI version
  help                 Show this help

Examples:
  amacli health
  amacli auth login
  amacli auth complete
  amacli me
  amacli source list
  amacli source set-default lenny
  amacli language set zh
  amacli search --query "How does Lenny think about MVP scope?"
  amacli search --balanced-content-types --query "What does Lenny say about PM hiring?" --top-k 6
  amacli document --id 42
  amacli save-answer --question "What does Lenny say about PM hiring?" --answer-file ./answer.md --citations-file ./citations.json

Environment:
  AMA_BASE_URL         Default: https://askmeanything.pro
  AMA_API_KEY          Optional when config.json already contains a key
  AMA_CONFIG_PATH      Optional path for amacli config.json
  AMA_HTTP_TIMEOUT     Optional HTTP timeout, default: 1m
`)
}
