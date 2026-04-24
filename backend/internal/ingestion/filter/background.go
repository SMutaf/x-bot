package filter

import "fmt"

func IsBackground(title string) bool {
	title = normalize(title)
	normalizedTitle := stripLeadingPunctuation(title)

	if wordCount(title) < 4 {
		fmt.Printf("[DEBUG] Background (kısa başlık): %.60s\n", title)
		return true
	}

	if hasTrailingQuestion(title) || hasTrailingQuestion(normalizedTitle) {
		fmt.Printf("[DEBUG] Background (soru işareti): %.60s\n", title)
		return true
	}

	for _, t := range emptyTitles {
		if hasPrefix(title, t) || title == t {
			fmt.Printf("[DEBUG] Background (boş başlık) '%s': %.60s\n", t, title)
			return true
		}
	}

	for _, p := range startsWithPatterns {
		if hasPrefix(title, p) || hasPrefix(normalizedTitle, p) {
			fmt.Printf("[DEBUG] Background (başta) '%s': %.60s\n", p, title)
			return true
		}
	}

	for _, p := range endsWithPatterns {
		if hasSuffix(title, p) {
			fmt.Printf("[DEBUG] Background (sonda) '%s': %.60s\n", p, title)
			return true
		}
	}

	for _, p := range titleOnlyPatterns {
		if contains(title, p) {
			fmt.Printf("[DEBUG] Background (başlıkta) '%s': %.60s\n", p, title)
			return true
		}
	}

	return false
}
