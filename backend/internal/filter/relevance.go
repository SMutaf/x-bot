package filter

func IsBreakingRelevant(text string) bool {
	if containsAny(text, rejectEntertainment) {
		return false
	}

	// Kesin içerik tipleri → red
	if containsAny(text, []string{
		"study suggests", "study finds", "review", "guide", "opinion",
		"how to", "what to know", "feature", "profile",
		"strategy to", "may prove",
	}) {
		return false
	}

	hasMajorActor := hasAnyMajorActor(text)
	hasAction := hasAnyActionVerb(text)

	// Türkiye + herhangi bir aksiyon → geç
	if containsAny(text, turkeyKeywords) && hasAction {
		return true
	}

	// Büyük aktör + aksiyon → geç
	if hasMajorActor && hasAction {
		return true
	}

	// Kritik güvenlik terimleri + büyük aktör → geç (aksiyon fiili olmasa bile)
	if containsAny(text, []string{
		"missile", "airstrike", "drone", "ceasefire", "hostage", "nuclear",
		"coup", "martial law", "earthquake", "terror", "terror alert", "gas facility",
		"landmine", "sanctions", "embargo",
	}) && hasMajorActor {
		return true
	}

	// Diplomatik hareketler — büyük aktör + hareket terimi → geç
	// "ABD'li heyet İslamabad'a indi", "müzakereler için uçak" gibi haberler
	if containsAny(text, []string{
		"negotiat", "ceasefire talks", "peace talks", "müzakere", "görüşme",
		"heyet", "delegation", "diplomat",
	}) && hasMajorActor {
		return true
	}

	return false
}

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

	// Türkiye keywordü varsa direkt geç
	if containsAny(text, turkeyKeywords) {
		return true
	}

	// Diplomatik/güvenlik olayları — aktör + aksiyon şartı
	criticalActors := []string{
		"iran", "israel", "trump", "erdogan", "fidan",
		"embassy", "ambassador", "diplomacy", "diplomatic", "foreign minister",
		"dışişleri", "savunma", "military", "ordu", "sınır",
		"pakistan", "islamabad", "hindistan", "india",
	}

	hasCriticalActor := containsAny(text, criticalActors)
	hasAction := hasAnyActionVerb(text)

	if hasCriticalActor && hasAction {
		return true
	}

	// Diplomatik hareket terimleri — aktör varsa aksiyon şartı arama
	if containsAny(text, []string{
		"müzakere", "negotiat", "heyet", "delegation", "ceasefire talks",
		"peace talks", "görüşme", "diplomat",
	}) && hasCriticalActor {
		return true
	}

	// Aksiyon fiili şartı aranmaksızın geçebilecek ağır terimler
	if containsAny(text, []string{
		"ateşkes", "patlama", "deprem", "saldırı", "seçim", "parliament",
		"gaz tesisi", "gas facility", "radar tesisi", "missile", "airstrike",
		"güvenlik konseyi", "nato summit", "bm", "landmine", "nükleer",
	}) {
		return true
	}

	return false
}
