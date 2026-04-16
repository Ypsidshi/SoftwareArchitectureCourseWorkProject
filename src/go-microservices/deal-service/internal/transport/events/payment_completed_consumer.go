package eventstransport

import (
	"context"
	"encoding/json"
	"log/slog"

	"coursework/deal-service/internal/service"
	"coursework/platform-common/pkg/events"
	"github.com/nats-io/nats.go"
)

type paymentHandler interface {
	HandlePaymentCompleted(ctx context.Context, event service.PaymentCompleted) error
}

func SubscribePaymentCompleted(conn *nats.Conn, logger *slog.Logger, handler paymentHandler) (*nats.Subscription, error) {
	return events.Subscribe(conn, "payment.completed", logger, func(env events.Envelope) error {
		var event service.PaymentCompleted
		if err := json.Unmarshal(env.Payload, &event); err != nil {
			return err
		}
		return handler.HandlePaymentCompleted(context.Background(), event)
	})
}
