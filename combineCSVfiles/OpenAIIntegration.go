package combinecsvfiles

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient wraps the OpenAI client and provides methods for our specific use cases
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAIClient instance
func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not found in environment variables")
	}

	client := openai.NewClient(apiKey)
	return &OpenAIClient{client: client}, nil
}

// SendPrompt sends a prompt to OpenAI and returns the response
func (c *OpenAIClient) SendPrompt(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a financial expert analyzing balance sheet data.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens: 1000,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// TestOpenAIConnection tests the connection to OpenAI
func TestOpenAIConnection() error {
	client, err := NewOpenAIClient()
	if err != nil {
		return fmt.Errorf("failed to create OpenAI client: %v", err)
	}

	ctx := context.Background()
	_, err = client.SendPrompt(ctx, "Test connection")
	if err != nil {
		return fmt.Errorf("failed to send test prompt: %v", err)
	}

	fmt.Println("Successfully connected to OpenAI!")
	return nil
}
