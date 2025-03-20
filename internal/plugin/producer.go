package plugin

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"not-leader-for-partition-example/config"
)

type Producer struct {
	client                      *kgo.Client
	adminClient                 *kadm.Client
	knownTopics                 kadm.TopicDetails
	timeout                     time.Duration
	updateMetadataBeforeProduce bool
	forceFlushAfterProduce      bool
}

func NewProducer(cfg *config.Config) (*Producer, error) {
	tlsDialer, err := NewTLSDialer(
		&tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		WithRootCAsPEMFromURL(CurlCACertPemURL))
	if err != nil {
		return nil, err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID("publisher-plugin"),
		kgo.WithLogger(NewKgoLogAdapter(kgo.LogLevelDebug)),
		kgo.Dialer(tlsDialer.DialContext),
		kgo.SASL(scram.Auth{
			User: cfg.Username,
			Pass: cfg.Password,
		}.AsSha256Mechanism()),
		kgo.ProducerBatchCompression(kgo.Lz4Compression(), kgo.ZstdCompression(), kgo.SnappyCompression(),
			kgo.GzipCompression(), kgo.NoCompression()),
		kgo.ProducerOnDataLossDetected(func(topic string, partition int32) {
			handler.Host.Log(api.LogLevelError,
				fmt.Sprintf("Data loss detected: topic=%s, partition=%d", topic, partition))
		}),
	}

	if recordRetries := cfg.RecordRetries; recordRetries > 0 {
		opts = append(opts, kgo.RecordRetries(recordRetries))
	}

	cli, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Producer{
		client:      cli,
		adminClient: kadm.NewClient(cli),
	}, nil
}

func (p *Producer) ProduceSync(ctx context.Context, record *kgo.Record) error {
	if err := p.upsertTopics(ctx, record.Topic); err != nil {
		return err
	}

	if p.updateMetadataBeforeProduce {
		p.client.ForceMetadataRefresh()
	}

	producectx := ctx
	if p.timeout > 0 {
		var cancel context.CancelFunc
		producectx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	if err := p.client.ProduceSync(producectx, record).FirstErr(); err != nil {
		return err
	}

	if !p.forceFlushAfterProduce {
		return nil
	}

	return p.client.Flush(ctx)
}
