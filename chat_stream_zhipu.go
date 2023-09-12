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

// CreateChatCompletionStream — API call to create a chat completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateChatCompletionStreamZhipu(
	ctx context.Context,
	request ChatCompletionRequest,
) (stream *ChatCompletionStreamZhipu, err error) {
	urlSuffix := "/sse-invoke"

	request.Stream = true
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
