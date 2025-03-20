package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	_ "github.com/stealthrocket/net/http"
	"github.com/stealthrocket/net/wasip1"
	"not-leader-for-partition-example/config"
	"not-leader-for-partition-example/internal/plugin"
)

func main() {}

//nolint:gochecknoinits // this is the only way to initialize the plugin in the wasm runtime
func init() {
	cfg, err := plugin.HydrateFromHostConfig[*config.Config](handler.Host.GetConfig())
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not load config: %v", err))
		os.Exit(1)
	}

	requiredFeatures := api.FeatureBufferRequest | api.FeatureBufferResponse
	if want, have := requiredFeatures, handler.Host.EnableFeatures(requiredFeatures); !have.IsEnabled(want) {
		handler.Host.Log(api.LogLevelError,
			fmt.Sprintf("unexpected features: want %s, have %s", want.String(), have.String()))
		os.Exit(1)
	}

	SetDefaultDNSResolver(cfg.DNSHost)
	SetDefaultHTTPTransport()

	plug, err := plugin.NewPlugin(cfg)
	if err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("Could not initialize plugin: %s", err.Error()))
		os.Exit(1)
	}

	handler.HandleRequestFn = plug.HandleRequest
	handler.HandleResponseFn = plug.HandleResponse
}

func SetDefaultHTTPTransport() {
	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		defaultTransport.DialContext = wasip1.DialContext
		defaultTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // ignoring G402 (CWE-295): Because there is no file mounted in the plugin by default, we configure insecureSkipVerify to avoid having to load rootCas
		}
	}
}

func SetDefaultDNSResolver(host string) {
	// Because there is no file mounted in the wasm binary by default, the default dns resolver needs to be configured
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, _ string, _ string) (net.Conn, error) {
			return (&wasip1.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext(ctx, "udp", net.JoinHostPort(host, "53"))
		},
	}
}
