package plugin

import "context"

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

	if _, err := p.adminClient.CreateTopics(ctx, 1, 3, nil, topic); err != nil {
		return err
	}

	return p.updateLocalTopics(ctx)
}

func (p *Producer) updateLocalTopics(ctx context.Context) error {
	topics, err := p.adminClient.ListTopics(ctx)
	if err != nil {
		return err
	}

	p.knownTopics = topics
	return nil
}
