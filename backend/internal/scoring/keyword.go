package scoring

import (
	"math"
	"strings"
	"unicode"
)

var lightKeywords = []string{
	"killed", "died", "arrested", "collapsed", "attacked",
	"struck", "seized", "resigned", "launched", "exploded",
	"sanctions", "evacuated", "invaded", "crashed",
	"öldü", "yaralandı", "tutuklandı", "saldırı", "patlama",
	"deprem", "çöktü", "istifa", "vuruldu",
	"faiz", "interest rate", "inflation", "enflasyon", "oil", "petrol", "lng",
	"fed", "gas facility", "energy", "radar", "resmi gazete", "dışişleri",
}

func KeywordScore(text string) float64 {
	lower := strings.ToLower(text)

	count := 0
	for _, kw := range lightKeywords {
		if strings.Contains(lower, kw) {
			count++
		}
	}

	if hasLargeNumber(lower) {
		count++
	}

	if count == 0 {
		return 0
	}

	return math.Min(float64(count*15), 100)
}

func hasLargeNumber(text string) bool {
	signals := []string{"billion", "trillion", "million", "milyar", "trilyon", "milyon", "%"}
	for _, s := range signals {
		if strings.Contains(text, s) {
			return true
		}
	}

	digits := 0
	for _, r := range text {
		if unicode.IsDigit(r) {
			digits++
			if digits >= 2 {
				return true
			}
		}
	}

	return false
}
