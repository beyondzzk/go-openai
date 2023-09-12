package openai

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	utils "github.com/beyondzzk/go-openai/internal"
)

var (
	headerData       = []byte("data: ")
	errorPrefix      = []byte(`data: {"error":`)
	zhipuHeaderId    = []byte("id:")
	zhipuHeaderEvent = []byte("event:")
	zhipuHeaderData  = []byte("data:")
	zhipuHeaderMeta  = []byte("meta:")
)

type streamable interface {
	ChatCompletionStreamResponse | CompletionResponse | ChatCompletionStreamZhipuResponse
}

type streamReader[T streamable] struct {
	emptyMessagesLimit uint
	isFinished         bool

	reader         *bufio.Reader
	response       *http.Response
	errAccumulator utils.ErrorAccumulator
	unmarshaler    utils.Unmarshaler
}

func (stream *streamReader[T]) Recv() (response T, err error) {
	if stream.isFinished {
		err = io.EOF
		return
	}

	response, err = stream.processLines()
	return
}

func (stream *streamReader[T]) processZhipuAILines() (T, error) {
	var (
		gotIdLine    bool
		gotDataLine  bool
		gotEventLine bool
		gotMetaLine  bool
	)

	tResponse := new(T)
	response, _ := any(tResponse).(*ChatCompletionStreamZhipuResponse)
	for {
		if gotIdLine && gotDataLine && gotEventLine {
			return *tResponse, nil
		}

		rawLine, readErr := stream.reader.ReadBytes('\n')
		noSpaceLine := bytes.TrimSpace(rawLine)
		if !gotIdLine && bytes.HasPrefix(noSpaceLine, zhipuHeaderId) {
			response.ID = string(bytes.TrimPrefix(noSpaceLine, zhipuHeaderId))
			gotIdLine = true
			if readErr == io.EOF {
				return *tResponse, nil
			}
			continue
		}

		if !gotEventLine && bytes.HasPrefix(noSpaceLine, zhipuHeaderEvent) {
			response.Event = string(bytes.TrimPrefix(noSpaceLine, zhipuHeaderEvent))
			gotEventLine = true
			if readErr == io.EOF {
				return *tResponse, nil
			}
			continue
		}

		if !gotDataLine && bytes.HasPrefix(noSpaceLine, zhipuHeaderData) {
			response.Data = string(bytes.TrimPrefix(noSpaceLine, zhipuHeaderData))
			gotDataLine = true
			if readErr == io.EOF {
				return *tResponse, nil
			}
			continue
		}

		if !gotMetaLine && bytes.HasPrefix(noSpaceLine, zhipuHeaderMeta) {
			if err := stream.unmarshaler.Unmarshal(bytes.TrimPrefix(noSpaceLine, zhipuHeaderMeta), &response.Meta); err != nil {

			}
			gotMetaLine = true
			if readErr == io.EOF {
				return *tResponse, nil
			}
			continue
		}

		if readErr != nil {
			if readErr == io.EOF {
				return *tResponse, nil
			}
			return *new(T), readErr
		}
	}
}

//nolint:gocognit
func (stream *streamReader[T]) processLines() (T, error) {
	var (
		emptyMessagesCount uint
		hasErrorPrefix     bool
		response           T
	)

	if _, ok := any(response).(ChatCompletionStreamZhipuResponse); ok {
		return stream.processZhipuAILines()
	}

	for {
		rawLine, readErr := stream.reader.ReadBytes('\n')
		if readErr != nil || hasErrorPrefix {
			respErr := stream.unmarshalError()
			if respErr != nil {
				return *new(T), fmt.Errorf("error, %w", respErr.Error)
			}
			return *new(T), readErr
		}

		noSpaceLine := bytes.TrimSpace(rawLine)
		if bytes.HasPrefix(noSpaceLine, errorPrefix) {
			hasErrorPrefix = true
		}
		if !bytes.HasPrefix(noSpaceLine, headerData) || hasErrorPrefix {
			if hasErrorPrefix {
				noSpaceLine = bytes.TrimPrefix(noSpaceLine, headerData)
			}
			writeErr := stream.errAccumulator.Write(noSpaceLine)
			if writeErr != nil {
				return *new(T), writeErr
			}
			emptyMessagesCount++
			if emptyMessagesCount > stream.emptyMessagesLimit {
				return *new(T), ErrTooManyEmptyStreamMessages
			}

			continue
		}

		noPrefixLine := bytes.TrimPrefix(noSpaceLine, headerData)
		if string(noPrefixLine) == "[DONE]" {
			stream.isFinished = true
			return *new(T), io.EOF
		}

		unmarshalErr := stream.unmarshaler.Unmarshal(noPrefixLine, &response)
		if unmarshalErr != nil {
			return *new(T), unmarshalErr
		}

		return response, nil
	}
}

func (stream *streamReader[T]) unmarshalError() (errResp *ErrorResponse) {
	errBytes := stream.errAccumulator.Bytes()
	if len(errBytes) == 0 {
		return
	}

	err := stream.unmarshaler.Unmarshal(errBytes, &errResp)
	if err != nil {
		errResp = nil
	}

	return
}

func (stream *streamReader[T]) Close() {
	stream.response.Body.Close()
}
