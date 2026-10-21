package plugin

import (
	"context"
	"fmt"
)

func (p *Producer) upsertTopics(ctx context.Context, topic string) error {
	if p.knownTopics.Has(topic) {
		return nil
	}

	if err := p.updateLocalTopics(ctx); err != nil {
		return err
	}

	if p.knownTopics.Has(topic) {
		return nil
	}

	_, err := p.adminClient.CreateTopics(ctx, defaultTopicPartitions, defaultTopicReplicationFactor, nil, topic)
	if err != nil {
		return fmt.Errorf("create topic %q: %w", topic, err)
	}

	return p.updateLocalTopics(ctx)
}

func (p *Producer) updateLocalTopics(ctx context.Context) error {
	topics, err := p.adminClient.ListTopics(ctx)
	if err != nil {
		return fmt.Errorf("list Kafka topics: %w", err)
	}

	p.knownTopics = topics
	return nil
}
