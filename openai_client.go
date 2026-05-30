package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type OpenAIClient struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

type ChatCompleter interface {
	Chat(ctx context.Context, messages []ChatMessage) (string, error)
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func NewOpenAIClientFromEnv() *OpenAIClient {
	return &OpenAIClient{
		BaseURL: strings.TrimRight(os.Getenv("OPENAI_BASE_URL"), "/"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *OpenAIClient) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if c == nil {
		return "", errors.New("OpenAI client is not configured")
	}
	if c.BaseURL == "" {
		return "", errors.New("OPENAI_BASE_URL is required")
	}
	if c.APIKey == "" {
		return "", errors.New("OPENAI_API_KEY is required")
	}
	if c.Model == "" {
		return "", errors.New("OPENAI_MODEL is required")
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.7,
	})
	if err != nil {
		return "", fmt.Errorf("encode chat completion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat completion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", c.sanitizeError(fmt.Errorf("chat completion request failed: %w", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", c.sanitizeError(fmt.Errorf("read chat completion response: %w", err))
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", c.sanitizeError(fmt.Errorf("decode chat completion response: status %d: %w", resp.StatusCode, err))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(respBody))
		if parsed.Error != nil && parsed.Error.Message != "" {
			message = parsed.Error.Message
		}
		return "", c.sanitizeError(fmt.Errorf("chat completion failed: status %d: %s", resp.StatusCode, message))
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", errors.New("chat completion returned empty content")
	}

	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func (c *OpenAIClient) sanitizeError(err error) error {
	message := err.Error()
	if c.APIKey != "" {
		message = strings.ReplaceAll(message, c.APIKey, "[redacted]")
	}
	return errors.New(message)
}
