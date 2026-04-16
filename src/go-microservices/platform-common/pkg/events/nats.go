package events

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Source     string          `json:"source"`
	TraceID    string          `json:"trace_id,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

func NewEnvelope(eventType, source, traceID string, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}

	return Envelope{
		EventID:    uuid.NewString(),
		EventType:  eventType,
		OccurredAt: time.Now().UTC(),
		Source:     source,
		TraceID:    traceID,
		Payload:    raw,
	}, nil
}

type Publisher interface {
	Publish(ctx context.Context, subject string, event Envelope) error
}

type NatsPublisher struct {
	conn *nats.Conn
}

func NewNATSPublisher(conn *nats.Conn) *NatsPublisher {
	return &NatsPublisher{conn: conn}
}

func (p *NatsPublisher) Publish(ctx context.Context, subject string, event Envelope) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := nats.NewMsg(subject)
	msg.Data = data
	return p.conn.PublishMsg(msg)
}

func ConnectNATS(url string) (*nats.Conn, error) {
	if url == "" {
		return nil, errors.New("nats url is empty")
	}
	return nats.Connect(url,
		nats.Name("coursework-ms"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second))
}

func Subscribe(conn *nats.Conn, subject string, logger *slog.Logger, handler func(Envelope) error) (*nats.Subscription, error) {
	return conn.Subscribe(subject, func(msg *nats.Msg) {
		var env Envelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			logger.Error("failed to decode event", slog.String("subject", subject), slog.String("error", err.Error()))
			return
		}
		if err := handler(env); err != nil {
			logger.Error("failed to process event", slog.String("event_type", env.EventType), slog.String("error", err.Error()))
		}
	})
}
