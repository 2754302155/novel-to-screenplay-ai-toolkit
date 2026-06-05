package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"gopkg.in/yaml.v3"
)

const defaultOpenAICompatibleBaseURL = "https://api.openai.com"

type OpenAICompatibleClient struct {
	config     ProviderConfig
	httpClient *http.Client
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewOpenAICompatibleClient(config ProviderConfig) (*OpenAICompatibleClient, error) {
	config.BaseURL = normalizeOpenAICompatibleEndpoint(config.BaseURL)
	config.Model = strings.TrimSpace(config.Model)
	config.APIKey = strings.TrimSpace(config.APIKey)

	if config.Model == "" {
		return nil, errors.New("model is required")
	}
	if config.APIKey == "" {
		return nil, errors.New("api key is required")
	}

	return &OpenAICompatibleClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}, nil
}

func normalizeOpenAICompatibleEndpoint(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return defaultOpenAICompatibleBaseURL + "/v1/chat/completions"
	}
	if strings.Contains(trimmed, "/v1/") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/chat/completions"
	}
	return trimmed + "/v1/chat/completions"
}

func (client *OpenAICompatibleClient) GenerateDraft(ctx context.Context, input DraftInput) (domain.ScreenplayDraft, error) {
	prompt := RenderScreenplayPrompt(input)
	content, err := client.chat(ctx, []chatMessage{
		{
			Role:    "system",
			Content: "你是专业的小说改编编剧助手。你必须只输出 YAML，不要输出 Markdown 代码围栏或解释。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	})
	if err != nil {
		return domain.ScreenplayDraft{}, err
	}

	normalizedYAML, err := normalizeDraftYAML(extractYAMLContent(content))
	if err != nil {
		return domain.ScreenplayDraft{}, err
	}

	var draft domain.ScreenplayDraft
	if err := yaml.Unmarshal(normalizedYAML, &draft); err != nil {
		return domain.ScreenplayDraft{}, fmt.Errorf("parse ai yaml: %w", err)
	}

	return draft, nil
}

func (client *OpenAICompatibleClient) TestConnection(ctx context.Context) error {
	content, err := client.chat(ctx, []chatMessage{
		{Role: "system", Content: "你只需要回答 ok。"},
		{Role: "user", Content: "请回复 ok"},
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("empty ai response")
	}

	return nil
}

func (client *OpenAICompatibleClient) chat(ctx context.Context, messages []chatMessage) (string, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model:       client.config.Model,
		Messages:    messages,
		Temperature: 0.2,
		MaxTokens:   4096,
	})
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+client.config.APIKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))
	if err != nil {
		return "", err
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(responseBody, &completion); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if completion.Error != nil && completion.Error.Message != "" {
			return "", errors.New(completion.Error.Message)
		}
		return "", fmt.Errorf("ai request failed with status %d", response.StatusCode)
	}
	if len(completion.Choices) == 0 {
		return "", errors.New("ai response has no choices")
	}

	return completion.Choices[0].Message.Content, nil
}

func normalizeDraftYAML(content string) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return nil, fmt.Errorf("parse ai yaml: %w", err)
	}
	normalizeStringListFields(&root)
	normalized, err := yaml.Marshal(&root)
	if err != nil {
		return nil, fmt.Errorf("normalize ai yaml: %w", err)
	}
	return normalized, nil
}

func normalizeStringListFields(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			key := node.Content[index]
			value := node.Content[index+1]
			if isStringListField(key.Value) {
				node.Content[index+1] = stringListNode(value)
				continue
			}
			normalizeStringListFields(value)
		}
		return
	}
	for _, child := range node.Content {
		normalizeStringListFields(child)
	}
}

func isStringListField(key string) bool {
	switch key {
	case "aliases",
		"themes",
		"source_refs",
		"characters",
		"notes",
		"timeline",
		"foreshadowing",
		"unresolved_issues",
		"carry_forward_notes",
		"warnings",
		"human_review_required":
		return true
	default:
		return false
	}
}

func stringListNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return emptySequenceNode()
	}
	if node.Kind == yaml.SequenceNode {
		hasMappingChild := false
		for _, child := range node.Content {
			if child.Kind == yaml.MappingNode {
				hasMappingChild = true
				normalizeStringListFields(child)
			}
		}
		if hasMappingChild {
			return node
		}
		for index, child := range node.Content {
			if child.Kind != yaml.ScalarNode || child.Tag != "!!str" {
				node.Content[index] = scalarStringNode(nodeValueAsString(child))
			}
		}
		return node
	}
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" && node.Value == "false" {
		return emptySequenceNode()
	}
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: []*yaml.Node{scalarStringNode(nodeValueAsString(node))},
	}
}

func emptySequenceNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Content: []*yaml.Node{}}
}

func scalarStringNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

func nodeValueAsString(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" && node.Value == "true" {
		return "AI 标记需要人工复核。"
	}
	if node.Value != "" {
		return node.Value
	}
	return "AI 输出结构需人工确认。"
}

func extractYAMLContent(content string) string {
	trimmed := strings.TrimSpace(content)
	if fenced := extractFencedYAML(trimmed); fenced != "" {
		return fenced
	}
	if index := strings.Index(trimmed, "schema_version:"); index >= 0 {
		return strings.TrimSpace(trimmed[index:])
	}
	if index := strings.Index(trimmed, "schema_version："); index >= 0 {
		return strings.TrimSpace(trimmed[index:])
	}
	return stripYAMLFences(trimmed)
}

func extractFencedYAML(content string) string {
	start := strings.Index(content, "```")
	if start < 0 {
		return ""
	}
	afterStart := content[start+3:]
	lineEnd := strings.Index(afterStart, "\n")
	if lineEnd < 0 {
		return ""
	}
	info := strings.TrimSpace(strings.ToLower(afterStart[:lineEnd]))
	if info != "" && info != "yaml" && info != "yml" {
		return ""
	}
	body := afterStart[lineEnd+1:]
	end := strings.Index(body, "```")
	if end < 0 {
		return strings.TrimSpace(body)
	}
	return strings.TrimSpace(body[:end])
}

func stripYAMLFences(content string) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```yaml")
	trimmed = strings.TrimPrefix(trimmed, "```yml")
	trimmed = strings.TrimPrefix(trimmed, "```YAML")
	trimmed = strings.TrimPrefix(trimmed, "```YML")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}
