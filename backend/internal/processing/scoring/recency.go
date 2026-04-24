package scoring

import "time"

func RecencyScore(publishedAt time.Time) float64 {
	if publishedAt.IsZero() {
		return 0
	}

	diff := time.Since(publishedAt).Minutes()
	switch {
	case diff < 5:
		return 100
	case diff < 15:
		return 85
	case diff < 30:
		return 70
	case diff < 60:
		return 55
	case diff < 120:
		return 35
	case diff < 240:
		return 22
	default:
		return 8
	}
}
