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
	disabled bool
	producer *Producer
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
		handleResponseErr(resp, fmt.Sprintf("failed to copy request body: %s", err.Error()))
		return false, 0
	}

	if err := p.producer.ProduceSync(context.Background(), &kgo.Record{
		Topic: "request",
		Value: body,
	}); err != nil {
		handleResponseErr(resp, fmt.Sprintf("failed to produce request record: %s", err.Error()))
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
		handleResponseErr(resp, fmt.Sprintf("failed to copy response body: %s", err.Error()))
		return
	}

	if err := p.producer.ProduceSync(context.Background(), &kgo.Record{
		Topic: "response",
		Value: body,
	}); err != nil {
		handleResponseErr(resp, fmt.Sprintf("failed to produce response record: %s", err.Error()))
		return
	}
}

func copyBody(body api.Body, writeBackBody bool) ([]byte, error) {
	var buf bytes.Buffer
	_, err := body.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	bs := buf.Bytes()
	if writeBackBody {
		body.Write(bs)
	}

	return bs, nil
}

func handleResponseErr(resp api.Response, errStr string) {
	resp.SetStatusCode(http.StatusInternalServerError)
	resp.Body().WriteString(errStr)
	handler.Host.Log(api.LogLevelError, errStr)
}
