package models

import "time"

type EditorialDecisionType string

const (
	DecisionPublish EditorialDecisionType = "PUBLISH"
	DecisionReject  EditorialDecisionType = "REJECT"
	DecisionReview  EditorialDecisionType = "REVİEW"
)

type ApprovalStatus string

const (
	ApprovalAutoApproved ApprovalStatus = "AUTO_APPROVED"
	ApprovalPending      ApprovalStatus = "PENDING_HUMAN_REVIEW"
	ApprovalApproved     ApprovalStatus = "APPROVED"
	ApprovalRejected     ApprovalStatus = "REJECTED"
)

type EditorialDecision struct {
	ID              string
	NewsID          string
	Decision        EditorialDecisionType
	RejectReason    string
	NewsType        string
	Sentiment       string
	Hook            string
	Summary         string
	Importance      string
	SourceLine      string
	ApprovalStatus  ApprovalStatus
	ApprovalChannel string
	CreatedAt       time.Time
}
