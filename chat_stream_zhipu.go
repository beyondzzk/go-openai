package openai

import (
	"context"
	"net/http"
)

type ChatCompletionMeta struct {
	Usage Usage `json:"usage"`
}

type ChatCompletionStreamZhipuResponse struct {
	ID    string             `json:"id"`
	Event string             `json:"event"`
	Data  string             `json:"data"`
	Meta  ChatCompletionMeta `json:"meta,omitempty"`
}

// ChatCompletionStream
// Note: Perhaps it is more elegant to abstract Stream using generics.
type ChatCompletionStreamZhipu struct {
	*streamReader[ChatCompletionStreamZhipuResponse]
}

// ChatCompletionRequest represents a request structure for chat completion API.
type ChatCompletionZhipuRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"prompt"`
	Temperature float32                 `json:"temperature,omitempty"`
	TopP        float32                 `json:"top_p,omitempty"`
}

// CreateChatCompletionStream — API call to create a chat completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateChatCompletionStreamZhipu(
	ctx context.Context,
	request ChatCompletionZhipuRequest,
) (stream *ChatCompletionStreamZhipu, err error) {
	urlSuffix := "/sse-invoke"

	req, err := c.newRequest(ctx, http.MethodPost, c.fullURL(urlSuffix, request.Model), withBody(request))
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestStream[ChatCompletionStreamZhipuResponse](c, req)
	if err != nil {
		return
	}
	stream = &ChatCompletionStreamZhipu{
		streamReader: resp,
	}
	return
}
