//go:generate curl -o cacert.pem https://curl.se/ca/cacert.pem
package plugin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/stealthrocket/net/wasip1"
)

const (
	CurlCACertPemURL      = "https://curl.se/ca/cacert.pem"
	defaultPEMLoadTimeout = 5 * time.Second
)

type Option func(*tls.Config)

type TLSDialer struct {
	config *tls.Config
}

func NewTLSDialer(config *tls.Config, opts ...Option) (*TLSDialer, error) {
	if !config.InsecureSkipVerify && (config.MinVersion == 0 || (config.MaxVersion != 0 && config.MaxVersion < config.MinVersion)) {
		return nil, errors.New("insecure configuration of TLS connection settings (G402, CWE-295): see https://deepsource.com/directory/go/issues/GSC-G402 for details")
	}

	for _, opt := range opts {
		opt(config)
	}

	return &TLSDialer{
		config: config.Clone(),
	}, nil
}

func (d *TLSDialer) DialContext(ctx context.Context, network string, address string) (net.Conn, error) {
	conn, err := wasip1.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	// one of InsecureSkipVerify or ServerName must be set, otherwise the connection will fail. If InsecureSkipVerify
	// is not set, try to set ServerName to the host part of the address
	if !d.config.InsecureSkipVerify && d.config.ServerName == "" {
		host, _, err := net.SplitHostPort(address)
		if err == nil {
			d.config.ServerName = host
		}
	}

	tlsConn := tls.Client(conn, d.config)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		if err := conn.Close(); err != nil {
			handler.Host.Log(api.LogLevelError, fmt.Sprintf("failed to close connection: %s", err.Error()))
		}

		return nil, err
	}

	return tlsConn, nil
}

func WithRootCAsPEMFromURL(url string) Option {
	return func(c *tls.Config) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultPEMLoadTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		if err != nil {
			panic(fmt.Errorf("failed to create request: %w", err))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(fmt.Errorf("failed to download CA bundle: %w", err))
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				handler.Host.Log(api.LogLevelError, fmt.Sprintf("failed to close response body: %s", err.Error()))
			}
		}()

		pemData, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(fmt.Errorf("failed to read CA bundle response: %w", err))
		}

		if c.RootCAs == nil {
			c.RootCAs = x509.NewCertPool()
		}

		if ok := c.RootCAs.AppendCertsFromPEM(pemData); !ok {
			panic(errors.New("failed to parse root certificate"))
		}
	}
}
