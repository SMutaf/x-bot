package scoring

type ScoreBreakdown struct {
	Cluster float64
	Recency float64
	Burst   float64
	Keyword float64
	Final   int
}
