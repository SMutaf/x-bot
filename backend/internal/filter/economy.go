package filter

import "strings"

func IsEconomyRelevant(text string) bool {
	// Kişisel finans içerikleri kesin reddet
	if containsAny(text, personalFinanceKeywords) {
		return false
	}
	// Türkiye veya global kritik etki VAR mı?
	hasTurkey := hasTurkeyImpact(text)
	hasGlobal := hasGlobalCriticalImpact(text)

	if !hasTurkey && !hasGlobal {
		return false
	}
	// Zayıf içerik ama kritik bir kurum söz konusuysa geç
	if isWeakEconomyContent(text) && !hasTurkey {
		return false
	}
	return true
}

func hasTurkeyImpact(text string) bool {
	keywords := []string{
		"türkiye", "turkey", "tcmb", "merkez bankası",
		"bist", "borsa istanbul", "lira", "try",
		"enflasyon türkiye",
	}

	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func hasGlobalCriticalImpact(text string) bool {
	keywords := []string{
		"fed",
		"ecb",
		"interest rate decision",
		"rate hike",
		"rate cut",
		"oil surge",
		"oil shock",
		"energy crisis",
		"financial crisis",
		"market crash",
		"global markets",
	}

	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

func isWeakEconomyContent(text string) bool {
	patterns := []string{
		"analysis",
		"we're watching",
		"outlook",
		"forecast",
		"strategist says",
		"commentary",
		"opinion",
		"here are",
		"this week",
	}

	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}
