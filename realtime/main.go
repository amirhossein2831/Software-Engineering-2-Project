package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"realtime/internal/config"
	"realtime/internal/events"
	"realtime/internal/hub"
	"realtime/internal/logger"
	"realtime/internal/server"
)

type envelope struct {
	Type    string    `json:"type"`
	EventID uuid.UUID `json:"event_id"`
	UserID  uuid.UUID `json:"user_id"`
}

func main() {
	log := logger.New("realtime")

	h := hub.New(config.GetInt("CLIENT_BUFFER", 64))

	topics := strings.Split(config.Get("CONSUME_TOPICS", "reservation.events,checkout.events,ticketing.events"), ",")
	group := config.Get("CONSUMER_GROUP", "realtime-"+uuid.NewString())
	consumer := events.NewConsumer(config.Get("KAFKA_BROKERS", ""), topics, group)

	go func() {
		if err := consumer.Run(context.Background(), func(ctx context.Context, key, value []byte) error {
			route(h, value)
			return nil
		}); err != nil {
			log.Error("consumer stopped", "err", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.ServeWS(h))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := ":" + config.Get("REALTIME_PORT", "8087")
	log.Info("realtime listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}

func route(h *hub.Hub, value []byte) {
	var ev envelope
	if err := json.Unmarshal(value, &ev); err != nil {
		return
	}
	if strings.HasPrefix(ev.Type, "seat.") {
		if ev.EventID != uuid.Nil {
			h.Broadcast("event:"+ev.EventID.String(), value)
		}
		return
	}
	if ev.UserID != uuid.Nil {
		h.Broadcast("user:"+ev.UserID.String(), value)
	}
}
