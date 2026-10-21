package plugin

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/twmb/franz-go/pkg/kgo"

	"not-leader-for-partition-example/config"
)

type Plugin struct {
	producer *Producer
	disabled bool
}

func NewPlugin(cfg *config.Config) (*Plugin, error) {
	producer, err := NewProducer(cfg)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		disabled: cfg.Disabled,
		producer: producer,
	}, nil
}

func (p *Plugin) HandleRequest(req api.Request, resp api.Response) (bool, uint32) {
	if p.disabled {
		return true, 0
	}

	body, err := copyBody(req.Body(), true)
	if err != nil {
		handleResponseErr(resp, fmt.Errorf("failed to copy request body: %w", err))
		return false, 0
	}

	err = p.producer.ProduceSync(context.Background(), &kgo.Record{
		Topic: "request",
		Value: body,
	})
	if err != nil {
		handleResponseErr(resp, fmt.Errorf("failed to produce request record: %w", err))
		return false, 0
	}

	return true, 0
}

func (p *Plugin) HandleResponse(_ uint32, _ api.Request, resp api.Response, _ bool) {
	if p.disabled {
		return
	}

	body, err := copyBody(resp.Body(), false)
	if err != nil {
		handleResponseErr(resp, fmt.Errorf("failed to copy response body: %w", err))
		return
	}

	err = p.producer.ProduceSync(context.Background(), &kgo.Record{
		Topic: "response",
		Value: body,
	})
	if err != nil {
		handleResponseErr(resp, fmt.Errorf("failed to produce response record: %w", err))
		return
	}
}

func copyBody(body api.Body, writeBackBody bool) ([]byte, error) {
	var buf bytes.Buffer
	_, err := body.WriteTo(&buf)
	if err != nil {
		return nil, fmt.Errorf("write body: %w", err)
	}

	bs := buf.Bytes()
	if writeBackBody {
		body.Write(bs)
	}

	return bs, nil
}

func handleResponseErr(resp api.Response, err error) {
	resp.SetStatusCode(http.StatusInternalServerError)
	resp.Body().WriteString(err.Error())
	handler.Host.Log(api.LogLevelError, err.Error())
}
