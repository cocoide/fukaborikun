package gateway

import (
	"context"
	"os"

	"github.com/sashabaranov/go-openai"
)

type OpenAIGateway interface {
	GetAnswerFromPrompt(prompt string) (string, error)
	AsyncGetAnswerFromPrompt(prompt string) <-chan string
}

type openAIGateway struct {
	client *openai.Client
	ctx    context.Context
}

func NewOpenAIGateway(ctx context.Context) OpenAIGateway {
	OPENAI_SECRET := os.Getenv("OPENAI_SECRET")
	client := openai.NewClient(OPENAI_SECRET)
	return &openAIGateway{client: client, ctx: ctx}
}

func (og *openAIGateway) GetAnswerFromPrompt(prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}
	res, err := og.client.CreateChatCompletion(og.ctx, req)
	if err != nil {
		return "", err
	}
	answer := res.Choices[0].Message.Content
	return answer, nil
}

func (og *openAIGateway) AsyncGetAnswerFromPrompt(prompt string) <-chan string {
	responseCh := make(chan string, 1)

	go func() {
		answer, _ := og.GetAnswerFromPrompt(prompt)
		responseCh <- answer
	}()

	return responseCh
}
