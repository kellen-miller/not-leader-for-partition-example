package config

import "time"

type Config struct {
	Disabled                    bool          `json:"disabled"`
	DNSHost                     string        `json:"dnsHost"`
	Username                    string        `json:"username"`
	Password                    string        `json:"password"`
	Brokers                     []string      `json:"brokers"`
	ProduceTimeout              time.Duration `json:"produceTimeout"`
	RecordRetries               int           `json:"recordRetries"`
	UpdateMetadataBeforeProduce bool          `json:"updateMetadataBeforeProduce"`
	FlushAfterProduce           bool          `json:"forceFlushAfterProduce"`
}
