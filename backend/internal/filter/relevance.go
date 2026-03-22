package filter

func IsBreakingRelevant(text string) bool {
	if containsAny(text, rejectEntertainment) {
		return false
	}

	if containsAny(text, []string{
		"study suggests", "study finds", "review", "guide", "analysis", "opinion",
		"how to", "what to know", "feature", "profile", "interview",
		"strategy to", "hopes", "whether to", "considers", "could", "may prove",
	}) {
		return false
	}

	hasMajorActor := false
	for actor := range majorActors {
		if contains(text, actor) {
			hasMajorActor = true
			break
		}
	}

	hasAction := false
	for verb := range actionVerbs {
		if contains(text, verb) {
			hasAction = true
			break
		}
	}

	if containsAny(text, turkeyKeywords) && hasAction {
		return true
	}

	if hasMajorActor && hasAction {
		return true
	}

	if containsAny(text, []string{
		"missile", "airstrike", "drone", "ceasefire", "hostage", "nuclear",
		"coup", "martial law", "earthquake", "terror", "terror alert", "gas facility",
	}) && hasMajorActor {
		return true
	}

	return false
}

/*func IsEconomyRelevant(text string) bool {
	if containsAny(text, personalFinanceKeywords) {
		return false
	}
	if containsAny(text, rejectEntertainment) {
		return false
	}
	if containsAny(text, []string{
		"opinion", "analysis", "review", "guide", "what to know", "feature",
		"healthy returns", "for consumers", "jim cramer", "wealth management push",
	}) {
		return false
	}

	if !containsAny(text, economyTerms) {
		return false
	}

	if !containsAny(text, strongEconomyTerms) && !containsAny(text, []string{
		"stocks decline", "prices surge", "fuel prices", "oil risk",
		"gas facility", "market volatility", "financial stability",
		"war fuels energy worries", "spiking oil prices",
	}) {
		return false
	}

	return true
}*/

func IsTechRelevant(text string) bool {
	if containsAny(text, rejectReviewPatterns) {
		return false
	}
	if containsAny(text, []string{
		"best deals", "discount", "sale", "robot vacuum", "smart plug", "tablet review",
	}) {
		return false
	}
	return containsAny(text, techTerms)
}

func IsGeneralRelevant(text string) bool {
	if containsAny(text, rejectEntertainment) {
		return false
	}

	if containsAny(text, turkeyKeywords) {
		return true
	}

	if containsAny(text, []string{
		"embassy", "ambassador", "diplomacy", "diplomatic", "foreign minister",
		"dışişleri", "savunma", "güvenlik", "military", "ordu", "sınır",
		"ateşkes", "patlama", "deprem", "saldırı", "seçim", "parliament",
		"gaz tesisi", "gas facility", "radar tesisi", "missile", "airstrike",
		"iran", "israel", "trump", "erdogan", "fidan",
	}) {
		return true
	}

	return false
}
