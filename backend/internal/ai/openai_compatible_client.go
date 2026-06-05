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

const defaultOpenAICompatibleBaseURL = "https://api.openai.com/v1"

type OpenAICompatibleClient struct {
	config     ProviderConfig
	httpClient *http.Client
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
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
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.Model = strings.TrimSpace(config.Model)
	config.APIKey = strings.TrimSpace(config.APIKey)

	if config.BaseURL == "" {
		config.BaseURL = defaultOpenAICompatibleBaseURL
	}
	if config.Model == "" {
		return nil, errors.New("model is required")
	}
	if config.APIKey == "" {
		return nil, errors.New("api key is required")
	}

	return &OpenAICompatibleClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
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

	var draft domain.ScreenplayDraft
	if err := yaml.Unmarshal([]byte(stripYAMLFences(content)), &draft); err != nil {
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
	})
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.config.BaseURL+"/chat/completions", bytes.NewReader(body))
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

func stripYAMLFences(content string) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```yaml")
	trimmed = strings.TrimPrefix(trimmed, "```yml")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}
