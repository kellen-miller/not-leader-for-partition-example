package config

import "time"

type Config struct {
	DNSHost                     string        `json:"dns_host"`
	Username                    string        `json:"username"`
	Password                    string        `json:"password"`
	Brokers                     []string      `json:"brokers"`
	ProduceTimeout              time.Duration `json:"produce_timeout"`
	RecordRetries               int           `json:"record_retries"`
	Disabled                    bool          `json:"disabled"`
	UpdateMetadataBeforeProduce bool          `json:"update_metadata_before_produce"`
	FlushAfterProduce           bool          `json:"force_flush_after_produce"`
}
