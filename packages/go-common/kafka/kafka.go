package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/events"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client *kgo.Client
}

func NewProducer(brokers []string) (*Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no kafka brokers configured")
	}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.MaxBufferedRecords(2000),
		kgo.ProducerLinger(10*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}
	return &Producer{client: client}, nil
}

func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("kafka producer is not configured")
	}
	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	return p.client.ProduceSync(ctx, rec).FirstErr()
}

func (p *Producer) Close() {
	if p != nil && p.client != nil {
		p.client.Close()
	}
}

type Handler func(ctx context.Context, env events.Envelope) error

// RunConsumer consumes until ctx is cancelled. Offsets are committed only after
// a handler succeeds; poison envelopes are logged and skipped so one bad
// message cannot stall the demo pipeline.
func RunConsumer(ctx context.Context, brokers []string, group string, topics []string, logger *slog.Logger, handler Handler) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers configured")
	}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
		kgo.SessionTimeout(15*time.Second),
		kgo.HeartbeatInterval(3*time.Second),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	logger.Info("kafka_consumer_started", "group", group, "topics", topics)
	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				if e.Err == context.Canceled {
					return nil
				}
				logger.Error("kafka_fetch_error", "topic", e.Topic, "partition", e.Partition, "error", e.Err.Error())
			}
		}

		var commit []*kgo.Record
		var handlerErr error
		iter := fetches.RecordIter()
		for !iter.Done() {
			rec := iter.Next()
			env, err := events.Parse(rec.Value)
			if err != nil {
				logger.Error("kafka_poison_message", "topic", rec.Topic, "offset", rec.Offset, "error", err.Error())
				commit = append(commit, rec)
				continue
			}
			if err := handler(ctx, env); err != nil {
				handlerErr = err
				logger.Error("kafka_handler_failed", "topic", rec.Topic, "type", env.Type, "event_id", env.ID, "error", err.Error())
				break
			}
			commit = append(commit, rec)
		}
		if len(commit) > 0 {
			if err := client.CommitRecords(ctx, commit...); err != nil {
				logger.Error("kafka_commit_failed", "error", err.Error())
			}
		}
		if handlerErr != nil {
			return handlerErr
		}
	}
}
