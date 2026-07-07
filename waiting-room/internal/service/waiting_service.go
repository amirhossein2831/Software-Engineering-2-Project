package service

import (
	"context"
	"time"

	"waitingroom/internal/queue"
	"waitingroom/internal/token"
)

type WaitingService struct {
	queue        *queue.Queue
	signer       *token.Signer
	admissionTTL time.Duration
}

func NewWaitingService(q *queue.Queue, signer *token.Signer, admissionTTL time.Duration) *WaitingService {
	return &WaitingService{queue: q, signer: signer, admissionTTL: admissionTTL}
}

type JoinResult struct {
	Status   string `json:"status"`
	Position int64  `json:"position"`
}

type StatusResult struct {
	Status         string `json:"status"`
	Position       int64  `json:"position,omitempty"`
	AdmissionToken string `json:"admission_token,omitempty"`
}

func (s *WaitingService) Join(ctx context.Context, eventID, userID string) (*JoinResult, error) {
	position, err := s.queue.Join(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if position == 0 {
		return &JoinResult{Status: "admitted", Position: 0}, nil
	}
	return &JoinResult{Status: "waiting", Position: position}, nil
}

func (s *WaitingService) Status(ctx context.Context, eventID, userID string) (*StatusResult, error) {
	admitted, position, err := s.queue.Status(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}
	if admitted {
		return &StatusResult{
			Status:         "admitted",
			AdmissionToken: s.signer.Issue(eventID, userID, s.admissionTTL),
		}, nil
	}
	return &StatusResult{Status: "waiting", Position: position}, nil
}

func (s *WaitingService) SetRate(ctx context.Context, eventID string, rate float64) error {
	return s.queue.SetRate(ctx, eventID, rate)
}
