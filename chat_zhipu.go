package openai

import (
	"context"
	"net/http"
)

type ChatCompletionZhipuChoice struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionData struct {
	Choice     []ChatCompletionZhipuChoice `json:"choices"`
	RequestId  string                      `json:"request_id,omitempty"`
	TaskId     int64                       `json:"task_id,omitempty"`
	TaskStatus string                      `json:"task_status,omitempty"`
	Usage      Usage                       `json:"usage"`
}

// ChatCompletionResponse represents a response structure for chat completion API.
type ChatCompletionZhipuResponse struct {
	Code    string             `json:"code,omitempty"`
	Message string             `json:"msg,omitempty"`
	Success int64              `json:"success,omitempty"`
	Data    ChatCompletionData `json:"data,omitempty"`
}

// CreateChatCompletion — API call to Create a completion for the chat message.
func (c *Client) CreateChatCompletionZhipu(
	ctx context.Context,
	request ChatCompletionZhipuRequest,
) (response ChatCompletionZhipuResponse, err error) {
	urlSuffix := "/invoke"

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
