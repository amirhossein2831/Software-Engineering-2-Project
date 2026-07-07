package sender

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"

	"notification/internal/model"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(ctx context.Context, m Message) error
}

type EmailSender struct {
	addr string
	from string
}

func NewEmailSender(addr, from string) *EmailSender {
	return &EmailSender{addr: addr, from: from}
}

func (s *EmailSender) Send(ctx context.Context, m Message) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.from, m.To, m.Subject, m.Body)
	return smtp.SendMail(s.addr, nil, s.from, []string{m.To}, []byte(msg))
}

type SmsSender struct {
	log *slog.Logger
}

func NewSmsSender(log *slog.Logger) *SmsSender {
	return &SmsSender{log: log}
}

func (s *SmsSender) Send(ctx context.Context, m Message) error {
	s.log.Info("sms delivered", "to", m.To, "body", m.Body)
	return nil
}

type Registry struct {
	senders map[model.Channel]Sender
}

func NewRegistry(email, sms Sender) *Registry {
	return &Registry{senders: map[model.Channel]Sender{
		model.ChannelEmail: email,
		model.ChannelSMS:   sms,
	}}
}

func (r *Registry) For(channel model.Channel) (Sender, bool) {
	s, ok := r.senders[channel]
	return s, ok
}
