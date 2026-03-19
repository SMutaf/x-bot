package filter

import (
	"strings"
	"unicode"
)

var startsWithPatterns = []string{
	"why ", "what is ", "what are ", "what do ", "what does ",
	"what would ", "what will ", "what should ",
	"how ", "who is ", "who are ", "who was ", "who will ",
	"where is ", "where are ", "where was ",
	"when is ", "when will ", "when did ",
	"is ", "are ", "was ", "were ",
	"will ", "would ", "could ", "should ",
	"can ", "did ", "does ", "do ",
	"has ", "have ", "had ",
	"watch:", "explainer:", "analysis:", "opinion:",
	"timeline:", "profile:", "live:", "fact check:",
	"in pictures:", "in charts:", "comment:", "guide:",
}

var endsWithPatterns = []string{
	" live",
	"- live",
	"– live",
	"— live",
	" politics live",
	" war live",
	" live updates",
	" as it happened",
	"- latest",
	"– latest",
}

var titleOnlyPatterns = []string{
	" explained",
	" analysis",
	" opinion",
	"what you need to know",
	"everything you need",
	"a guide to",
	" live updates",
	"crisis live:",
	" politics live",
	" war live",
	"what to know",
	"review",
	"hands-on",
	"first look",
	"deals",
	" alınır mı",
	" inceleme",
	" karşılaştırma",
}

var emptyTitles = []string{
	"here's the latest",
	"morning news brief",
	"evening news brief",
	"news brief",
	"the latest",
	"top stories",
	"this week in",
	"coming up on",
	"live updates",
	"open interest",
	"wall street week",
}

var actionVerbs = map[string]int{
	"attack":   16,
	"strike":   16,
	"launch":   12,
	"kill":     18,
	"die":      18,
	"warn":     10,
	"sanction": 12,
	"ban":      10,
	"explode":  16,
	"crash":    16,
	"resign":   12,
	"arrest":   12,
	"close":    10,
	"halt":     10,
	"deploy":   12,
	"invade":   18,
	"bomb":     18,
	"shoot":    16,
	"hit":      12,
	"destroy":  16,
	"collapse": 16,
	"capture":  12,
	"seize":    12,
	"detain":   10,
	"raid":     12,
	"saldırı":  16,
	"deprem":   20,
	"patlama":  16,
	"öldü":     18,
	"yaralı":   12,
	"idam":     16,
	"ateşkes":  12,
	"vurdu":    16,
	"vuruldu":  16,
	"patladı":  16,
}

var majorActors = map[string]int{
	"america":       8,
	"united states": 8,
	"china":         8,
	"russia":        8,
	"iran":          10,
	"israel":        10,
	"turkey":        14,
	"türkiye":       14,
	"pakistan":      8,
	"japan":         6,
	"south korea":   6,
	"north korea":   8,
	"taiwan":        8,
	"ukraine":       8,
	"gaza":          8,
	"hamas":         8,
	"hezbollah":     8,
	"saudi":         6,
	"nato":          8,
	"pentagon":      8,
	"white house":   8,
	"kremlin":       8,
	"moscow":        8,
	"washington":    8,
	"beijing":       8,
	"tehran":        8,
	"erdogan":       12,
	"trump":         12,
	"beyaz saray":   8,
	"çin":           8,
	"rusya":         8,
	"ukrayna":       8,
	"fransa":        6,
	"almanya":       6,
	"avrupa":        6,
	"tahran":        8,
	"moskova":       8,
	"ab":            8,
	"abd":           8,
	"fidan":         10,
}

var shortActors = map[string]int{
	"us":  8,
	"un":  8,
	"eu":  6,
	"uk":  6,
	"fed": 10,
	"imf": 10,
}

var turkeyKeywords = []string{
	"turkey", "türkiye", "istanbul", "ankara", "tbmm", "cumhurbaşkanı", "bakanlık",
	"resmi gazete", "savunma sanayii", "dışişleri", "içişleri", "meclis", "marmara",
}

var economyTerms = []string{
	"fed", "faiz", "interest rate", "inflation", "enflasyon", "cpi", "ppi",
	"central bank", "merkez bankası", "banka", "bank", "oil", "petrol",
	"gas", "lng", "brent", "barrel", "kur", "dolar", "euro", "lira",
	"bond", "tahvil", "yield", "treasury", "market", "borsa", "bist",
	"stock", "shares", "equity", "trade", "tariff", "tariffs", "sanction",
	"gdp", "işsizlik", "unemployment", "exports", "imports", "ithalat", "ihracat",
	"gold", "altın", "energy", "enerji", "opec", "fitch", "moody", "s&p",
}

var strongEconomyTerms = []string{
	"fed", "interest rate", "faiz", "inflation", "enflasyon", "central bank",
	"merkez bankası", "oil", "petrol", "lng", "brent", "barrel", "yield",
	"treasury", "bond", "kur", "dolar", "euro", "energy", "enerji", "opec",
	"market crash", "resesyon", "recession", "gas facility", "fuel prices",
}

var techTerms = []string{
	"ai", "yapay zeka", "openai", "chatgpt", "anthropic", "google", "meta", "microsoft",
	"apple", "nvidia", "amd", "intel", "tesla", "cyber", "hack", "hacker", "data breach",
	"security", "güvenlik", "chip", "semiconductor", "startup", "robot", "robotics",
	"software", "app", "uygulama", "iphone", "android", "gemini", "claude",
	"deepmind", "api", "cloud", "aws", "azure", "cloudflare", "fbi is buying americans’ location data",
	"fbi is buying americans' location data",
}

var rejectEntertainment = []string{
	"movie", "film", "concert", "festival", "celebrity", "toy", "podcast",
	"ice cream", "smelly", "ad falls flat", "viral toys", "bts comeback concert",
}

var rejectReviewPatterns = []string{
	"review", "hands-on", "first look", "best deals", "deal", "discount", "sale",
	"alınır mı", "inceleme", "karşılaştırma", "en iyi", "fiyatı", "kaç tl",
}

var personalFinanceKeywords = []string{
	"my husband", "my wife", "my partner", "my salary",
	"i'm retired", "i retired", "retire early", "retirement plan",
	"credit card", "credit score", "credit limit",
	"401(k)", "roth ira", "tax refund", "mortgage rate",
	"renting vs", "buying a home", "personal finance",
	"cost of living", "student loan", "pay off debt",
}

func normalize(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

func contains(text string, word string) bool {
	return strings.Contains(text, word)
}

func containsWord(text string, word string) bool {
	idx := strings.Index(text, word)
	if idx == -1 {
		return false
	}
	before := idx == 0 || text[idx-1] == ' '
	after := idx+len(word) == len(text) || text[idx+len(word)] == ' '
	return before && after
}

func containsAny(text string, words []string) bool {
	for _, w := range words {
		if contains(text, w) {
			return true
		}
	}
	return false
}

func hasNumber(text string) bool {
	for _, r := range text {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func stripLeadingPunctuation(text string) string {
	for _, sep := range []string{"': ", "\": ", "' ", "\" "} {
		if idx := strings.Index(text, sep); idx != -1 && idx < 60 {
			after := strings.TrimSpace(text[idx+len(sep):])
			if len(after) > 10 {
				return after
			}
		}
	}
	return text
}

func wordCount(text string) int {
	return len(strings.Fields(text))
}

func hasTrailingQuestion(text string) bool {
	return strings.HasSuffix(text, "?")
}

func hasPrefix(text, prefix string) bool {
	return strings.HasPrefix(text, prefix)
}

func hasSuffix(text, suffix string) bool {
	return strings.HasSuffix(text, suffix)
}
