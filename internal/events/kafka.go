package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

// Event represents a system event
type Event struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	HandleEvent(ctx context.Context, event *Event) error
}

// KafkaProducer handles event production
type KafkaProducer struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, logger *zap.Logger) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer: producer,
		logger:   logger,
	}, nil
}

// PublishEvent publishes an event to Kafka
func (p *KafkaProducer) PublishEvent(ctx context.Context, topic string, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// KafkaConsumer handles event consumption
type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handlers map[string]EventHandler
	logger   *zap.Logger
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers []string, groupID string, logger *zap.Logger) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		consumer: consumer,
		handlers: make(map[string]EventHandler),
		logger:   logger,
	}, nil
}

// RegisterHandler registers an event handler for a specific event type
func (c *KafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

// Start starts consuming events
func (c *KafkaConsumer) Start(ctx context.Context, topics []string) error {
	consumer := &consumerGroupHandler{
		handlers: c.handlers,
		logger:   c.logger,
	}

	for {
		err := c.consumer.Consume(ctx, topics, consumer)
		if err != nil {
			return fmt.Errorf("failed to consume messages: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handlers map[string]EventHandler
	logger   *zap.Logger
}

// Setup is called when the consumer group is set up
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called when the consumer group is torn down
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a claim
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.Error("Failed to unmarshal event",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int64("offset", msg.Offset))
			continue
		}

		handler, ok := h.handlers[event.Type]
		if !ok {
			h.logger.Warn("No handler registered for event type",
				zap.String("type", event.Type))
			continue
		}

		if err := handler.HandleEvent(session.Context(), &event); err != nil {
			h.logger.Error("Failed to handle event",
				zap.Error(err),
				zap.String("type", event.Type))
			continue
		}

		session.MarkMessage(msg, "")
	}

	return nil
}

// Common event types
const (
	EventUserCreated     = "user.created"
	EventUserUpdated     = "user.updated"
	EventUserDeleted     = "user.deleted"
	EventCredentialAdded = "credential.added"
	EventMFAMethodAdded  = "mfa.added"
) 