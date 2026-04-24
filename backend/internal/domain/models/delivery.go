package models

import "time"

type DeliveryTarget string

const (
	TargetTelegram DeliveryTarget = "TELEGRAM"
	TargetPanel    DeliveryTarget = "PANEL"
)

type RenderedContent struct {
	Target      DeliveryTarget
	Title       string
	Body        string
	SourceLine  string
	URL         string
	GeneratedAt time.Time
}

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "PENDING"
	DeliverySent    DeliveryStatus = "SENT"
	DeliveryFailed  DeliveryStatus = "FAILED"
)

type DeliveryRecord struct {
	ID         string
	DecisionID string
	Target     DeliveryTarget
	Status     DeliveryStatus
	Error      string
	SentAt     *time.Time
}
