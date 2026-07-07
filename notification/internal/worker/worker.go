package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"notification/internal/model"
	"notification/internal/repository"
	"notification/internal/sender"
)

type Worker struct {
	repo        *repository.NotificationRepo
	registry    *sender.Registry
	log         *slog.Logger
	maxAttempts int
	batch       int
	interval    time.Duration
}

func New(repo *repository.NotificationRepo, registry *sender.Registry, log *slog.Logger, maxAttempts, batch int, interval time.Duration) *Worker {
	return &Worker{repo: repo, registry: registry, log: log, maxAttempts: maxAttempts, batch: batch, interval: interval}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	rows, err := w.repo.PendingBatch(ctx, w.batch)
	if err != nil {
		w.log.Error("pending fetch failed", "err", err)
		return
	}
	for _, row := range rows {
		w.deliver(ctx, row)
	}
}

func (w *Worker) deliver(ctx context.Context, row model.NotificationOutbox) {
	snd, ok := w.registry.For(row.Channel)
	if !ok {
		_ = w.repo.MarkAttempt(ctx, row.ID, row.Attempts+1, model.StatusFailed)
		return
	}
	var payload map[string]any
	_ = json.Unmarshal([]byte(row.Payload), &payload)
	subject, body := render(row.Template, payload)

	if err := snd.Send(ctx, sender.Message{To: row.To, Subject: subject, Body: body}); err != nil {
		attempts := row.Attempts + 1
		status := model.StatusPending
		if attempts >= w.maxAttempts {
			status = model.StatusFailed
		}
		_ = w.repo.MarkAttempt(ctx, row.ID, attempts, status)
		w.log.Warn("delivery failed", "id", row.ID, "attempts", attempts, "err", err)
		return
	}
	_ = w.repo.MarkSent(ctx, row.ID)
}

func render(template string, p map[string]any) (subject, body string) {
	switch template {
	case "payment_succeeded":
		return "Your payment succeeded",
			fmt.Sprintf("Order %v is confirmed. Amount charged: %v.", p["order_id"], p["amount"])
	case "payment_failed":
		return "Payment could not be completed",
			fmt.Sprintf("Order %v failed and your seats were released. Please try again.", p["order_id"])
	case "ticket_issued":
		return "Your ticket is ready",
			fmt.Sprintf("Ticket %v for event %v, seat %v. QR code: %v", p["ticket_id"], p["event_id"], p["seat_id"], p["qr_hash"])
	default:
		return "Notification", ""
	}
}
