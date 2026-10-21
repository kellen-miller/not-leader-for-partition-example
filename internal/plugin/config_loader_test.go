//go:build !wasip1

package plugin

import (
	"encoding/json"
	"errors"
	"testing"

	"not-leader-for-partition-example/config"
)

func TestHydrateFromHostConfigAcceptsSnakeCaseKeys(t *testing.T) {
	cfg, err := HydrateFromHostConfig[config.Config](
		json.RawMessage(
			`{"dns_host":"127.0.0.11","produce_timeout":"10s","record_retries":"2","update_metadata_before_produce":"true","force_flush_after_produce":"true"}`,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.DNSHost != "127.0.0.11" || cfg.ProduceTimeout.String() != "10s" || cfg.RecordRetries != 2 {
		t.Fatalf("config values were not hydrated from snake_case keys: %+v", cfg)
	}

	if !cfg.UpdateMetadataBeforeProduce || !cfg.FlushAfterProduce {
		t.Fatalf("boolean config values were not hydrated from snake_case keys: %+v", cfg)
	}
}

func TestHydrateFromHostConfigAcceptsLegacyCamelCaseKeys(t *testing.T) {
	cfg, err := HydrateFromHostConfig[config.Config](
		json.RawMessage(
			`{"dnsHost":"127.0.0.11","produceTimeout":"10s","recordRetries":"2","updateMetadataBeforeProduce":"true","forceFlushAfterProduce":"true"}`,
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.DNSHost != "127.0.0.11" || cfg.ProduceTimeout.String() != "10s" || cfg.RecordRetries != 2 {
		t.Fatalf("config values were not hydrated from legacy camelCase keys: %+v", cfg)
	}

	if !cfg.UpdateMetadataBeforeProduce || !cfg.FlushAfterProduce {
		t.Fatalf("boolean config values were not hydrated from legacy camelCase keys: %+v", cfg)
	}
}

func TestHydrateFromHostConfigWrapsJSONErrors(t *testing.T) {
	_, err := HydrateFromHostConfig[config.Config](json.RawMessage(`{`))
	if err == nil {
		t.Fatal("expected invalid JSON to fail")
	}

	if _, ok := errors.AsType[*json.SyntaxError](err); !ok {
		t.Fatalf("expected wrapped JSON syntax error, got %v", err)
	}
}
