package filter

import "strings"

func IsEconomyRelevant(text string) bool {
	// Kişisel finans içerikleri kesin reddet
	if containsAny(text, personalFinanceKeywords) {
		return false
	}

	hasTurkey := hasTurkeyImpact(text)
	hasGlobal := hasGlobalCriticalImpact(text)

	if !hasTurkey && !hasGlobal {
		return false
	}

	// Zayıf içerik (outlook, commentary vb.) ama Türkiye etkisi varsa geç
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
		// Merkez bankaları & faiz
		"fed", "ecb", "interest rate decision", "rate hike", "rate cut",
		"federal reserve", "rate decision",

		// Enerji fiyat şokları
		"oil surge", "oil shock", "oil spike", "oil crash",
		"energy crisis", "energy shock",
		"gas shortage", "gas supply",

		// Deniz yolu & ticaret akışı — Hormuz / Süveyş kritik
		"hormuz", "strait of hormuz", "hürmüz",
		"suez", "red sea", "shipping halt", "shipping disruption",
		"shipping traffic", "trade route", "supply chain disruption",
		"port closure", "tanker",

		// Piyasa çöküşleri
		"financial crisis", "market crash", "market collapse",
		"stock market crash", "global markets",

		// Yaptırım & ticaret savaşı
		"sanctions", "embargo", "tariff war", "trade war",
		"export ban", "import ban",

		// IMF / Dünya Bankası kararları
		"imf bailout", "imf deal", "world bank",

		// Emtia şokları
		"commodity shock", "wheat shortage", "food crisis",
		"grain supply", "opec cut", "opec decision",
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
		"analysis", "we're watching", "outlook", "forecast",
		"strategist says", "commentary", "opinion",
		"here are", "this week",
	}
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}
