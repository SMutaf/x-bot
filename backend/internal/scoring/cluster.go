package scoring

func ClusterScore(clusterCount int) float64 {
	switch {
	case clusterCount >= 5:
		return 100
	case clusterCount >= 4:
		return 92
	case clusterCount >= 3:
		return 82
	case clusterCount >= 2:
		return 65
	default:
		return 0
	}
}
