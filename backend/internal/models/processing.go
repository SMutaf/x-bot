package models

type ProcessingStage string

const (
	StageIngested    ProcessingStage = "INGESTED"
	StageFiltered    ProcessingStage = "FILTERED"
	StageScored      ProcessingStage = "SCORED"
	StageClustered   ProcessingStage = "CLUSTERED"
	StageLLMAnalyzed ProcessingStage = "LLM_ANALYZED"
	StageApproved    ProcessingStage = "APPROVED"
	StageDelivered   ProcessingStage = "DELIVERED"
)

type FilterDecision struct {
	Passed bool
	Reason string
}

type ScoreBreakdown struct {
	Cluster float64
	Recency float64
	Burst   float64
	Keyword float64
	Final   int
	Boost   int
}

type ClusterInfo struct {
	ClusterKey   string
	ClusterCount int
	IsClustered  bool
}

type NewsEnvelope struct {
	News        RawNewsItem
	Stage       ProcessingStage
	Filter      FilterDecision
	Score       ScoreBreakdown
	Cluster     ClusterInfo
	NeedsReview bool
}

func NewEnvelope(news RawNewsItem) NewsEnvelope {
	return NewsEnvelope{
		News:  news,
		Stage: StageIngested,
		Filter: FilterDecision{
			Passed: true,
			Reason: "pending",
		},
		Score: ScoreBreakdown{},
		Cluster: ClusterInfo{
			ClusterKey:   "",
			ClusterCount: 1,
			IsClustered:  false,
		},
	}
}
