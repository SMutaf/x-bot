package scoring

import (
	"math"
	"strings"
)

type keywordTier struct {
	weight float64
	words  []string
}

var keywordTiers = []keywordTier{
	{
		weight: 26,
		words: []string{
			"deprem", "earthquake", "nükleer", "nuclear", "savaş", "war",
			"tsunami", "collapsed", "çöktü", "missile", "füze",
		},
	},
	{
		weight: 18,
		words: []string{
			"öldü", "killed", "died", "patlama", "explosion", "saldırı", "attack",
			"ateşkes", "ceasefire", "tutuklama", "tutuklandı", "arrested",
			"istifa", "resigned", "vuruldu", "struck", "seized", "invade", "invaded",
			"rehine", "hostage", "yangın", "wildfire", "sel", "flood",
		},
	},
	{
		weight: 12,
		words: []string{
			"faiz kararı", "interest rate", "rate hike", "rate cut", "sanctions",
			"embargo", "yaptırım", "iflas", "bankruptcy", "market crash",
			"borsa çöküşü", "fed", "tcmb", "seçim", "election", "referendum",
			"cumhurbaşkanı", "meclis", "tbmm", "anayasa", "gsyh", "gdp",
			"cari açık", "işsizlik", "unemployment", "default",
		},
	},
	{
		weight: 8,
		words: []string{
			"oil", "petrol", "enflasyon", "inflation", "dolar", "euro", "lira",
			"lng", "energy", "enerji", "launched", "duyurdu", "açıkladı",
			"borsa", "ihracat", "ithalat", "trump", "putin", "xi jinping",
			"nato", "israil", "israel", "iran", "gazze", "gaza", "ukrayna",
			"ukraine", "imf", "world bank", "opec", "oil price", "gold", "altın",
			"resmi gazete", "dışişleri", "ankara", "istanbul",
		},
	},
}

var turkeyPrimarySignals = []string{
	"türkiye", "turkey", "tcmb", "bist", "borsa istanbul",
	"lira", "try", "erdoğan", "erdogan", "ankara", "istanbul", "tbmm",
}

var turkeySecondarySignals = []string{
	"meclis", "cumhurbaşkanı", "cumhurbaskani", "merkez bankası", "merkez bankasi",
	"resmi gazete", "dışişleri", "disisleri", "izmir", "bursa",
}

var magnitudeSignals = []string{
	"billion", "trillion", "million", "milyar", "trilyon", "milyon",
	"basis points", "baz puan", "yüzde", "%", "bn", "mn",
}

var magnitudeMoveSignals = []string{
	"artış", "artis", "increase", "surge", "jump", "drop", "düşüş", "dusus",
	"fall", "plunge", "gain", "loss", "decline", "rally",
}

func KeywordScore(text string) float64 {
	lower := normalizeScoreText(text)

	total := 0.0
	matches := 0
	seen := make(map[string]struct{})

	for _, tier := range keywordTiers {
		for _, kw := range tier.words {
			if _, ok := seen[kw]; ok {
				continue
			}
			if !strings.Contains(lower, kw) {
				continue
			}

			seen[kw] = struct{}{}
			decay := math.Pow(0.78, float64(matches))
			total += tier.weight * decay
			matches++
		}
	}

	if total > 100 {
		return 100
	}

	return total
}

func TurkeyRelevanceScore(text string) float64 {
	lower := normalizeScoreText(text)

	score := 0.0
	for _, kw := range turkeyPrimarySignals {
		if strings.Contains(lower, kw) {
			score += 12
		}
	}

	for _, kw := range turkeySecondarySignals {
		if strings.Contains(lower, kw) {
			score += 6
		}
	}

	if score > 30 {
		return 30
	}

	return score
}

func MagnitudeScore(text string) float64 {
	lower := normalizeScoreText(text)

	hasMagnitude := false
	for _, kw := range magnitudeSignals {
		if strings.Contains(lower, kw) {
			hasMagnitude = true
			break
		}
	}

	if !hasMagnitude {
		return 0
	}

	for _, kw := range magnitudeMoveSignals {
		if strings.Contains(lower, kw) {
			return 18
		}
	}

	return 10
}

func normalizeScoreText(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}
