package policy

import (
	"strings"

	"github.com/SMutaf/twitter-bot/backend/internal/domain/models"
)

func IsCriticalEvent(env models.NewsEnvelope, _ CategoryPolicy) bool {
	text := strings.ToLower(env.News.Title + " " + env.News.Description)

	strongEvent := hasStrongAction(text) && (hasSecurityEvent(text) || hasEnergyShockEvent(text) || hasMacroEvent(text))

	if !strongEvent {
		if isWeakDiplomaticText(text) || isWeakSpeculativeText(text) {
			return false
		}
	}

	if hasMacroEvent(text) {
		return true
	}

	if hasOfficialAction(text) && (hasStrongAction(text) || hasNumberSignal(text) || hasStrongSecondarySignal(text)) {
		return true
	}

	if hasSecurityEvent(text) && (hasStrongAction(text) || hasNumberSignal(text) || hasStrongSecondarySignal(text)) {
		return true
	}

	if hasEnergyShockEvent(text) && (hasStrongAction(text) || hasNumberSignal(text) || hasStrongSecondarySignal(text)) {
		return true
	}

	return false
}

func isWeakDiplomaticText(text string) bool {
	weakPatterns := []string{
		"telefonda görüştü",
		"telefon görüşmesi",
		"görüştü",
		"görüştüler",
		"görüşme gerçekleştirdi",
		"görüş alışverişinde bulundu",
		"memnuniyet duydu",
		"memnun olduğunu belirtti",
		"kınadı",
		"tepki gösterdi",
		"açıklama yaptı",
		"değerlendirdi",
		"mesaj verdi",
		"çağrıda bulundu",
		"temasta bulundu",
		"bir araya geldi",
		"ile görüştü",
		"mevkidaşı ile görüştü",
		"övdü",
		"anlayış istedi",
	}

	for _, p := range weakPatterns {
		if strings.Contains(text, p) {
			return true
		}
	}

	return false
}

func isWeakSpeculativeText(text string) bool {
	weakSpeculativePatterns := []string{
		"olası",
		"ihtimal",
		"ankete göre",
		"anket",
		"destekliyor",
		"hazırlık yapıyoruz",
		"hazırlık yapıyor",
		"hazırlık yapacak",
		"iddiası",
		"iddia",
		"analiz etti",
		"analiz",
		"değerlendirme",
		"we're watching",
		"here are",
		"this week",
		"things we're watching",
	}

	for _, p := range weakSpeculativePatterns {
		if strings.Contains(text, p) {
			return true
		}
	}

	return false
}

func hasMacroEvent(text string) bool {
	macroKeywords := []string{
		"fed",
		"interest rate",
		"faiz",
		"policy rate",
		"merkez bankası",
		"central bank",
		"inflation",
		"enflasyon",
		"cpi",
		"ppi",
		"rate hike",
		"rate cut",
		"faizi sabit",
		"faiz kararı",
		"faizi değiştirmedi",
		"held rates",
		"raised rates",
		"cut rates",
	}

	for _, kw := range macroKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func hasSecurityEvent(text string) bool {
	securityKeywords := []string{
		"missile",
		"airstrike",
		"drone strike",
		"attack",
		"strike",
		"saldırı",
		"patlama",
		"füze",
		"drone",
		"iha",
		"siha",
		"earthquake",
		"deprem",
		"ceasefire",
		"ateşkes",
		"nuclear",
		"radar tesisi",
		"radar facility",
		"war",
		"savaş",
		"martial law",
		"hostage",
		"çıkarma yapmak",
		"hava saldırısı",
	}

	for _, kw := range securityKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func hasEnergyShockEvent(text string) bool {
	energyKeywords := []string{
		"oil",
		"petrol",
		"lng",
		"gas field",
		"gas facility",
		"energy facility",
		"energy infrastructure",
		"barrel",
		"brent",
		"hürmüz",
		"hormuz",
		"boğazı",
		"strait of hormuz",
		"refinery",
		"rafineri",
		"doğal gaz",
		"natural gas",
		"akaryakıt",
	}

	for _, kw := range energyKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func hasOfficialAction(text string) bool {
	officialKeywords := []string{
		"resmi gazete",
		"sanction",
		"yaptırım",
		"ban",
		"yasak",
		"martial law",
		"olağanüstü hal",
		"state of emergency",
		"seferberlik",
		"mobilization",
		"official statement",
		"resmi karar",
	}

	for _, kw := range officialKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func hasStrongAction(text string) bool {
	actionKeywords := []string{
		"killed",
		"dead",
		"died",
		"destroyed",
		"hit",
		"struck",
		"halted",
		"surged",
		"fell",
		"rose",
		"cut rates",
		"held rates",
		"raised rates",
		"launched",
		"bombed",
		"attacked",
		"collapsed",
		"executed",
		"öldü",
		"öldürüldü",
		"yaralandı",
		"yaralı",
		"vuruldu",
		"vurdu",
		"yükseldi",
		"geriledi",
		"sabit tuttu",
		"faizi değiştirmedi",
		"faizi artırdı",
		"faizi indirdi",
		"imha edildi",
		"saldırdı",
		"hedef aldı",
		"hayatını kaybetti",
	}

	for _, kw := range actionKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func hasNumberSignal(text string) bool {
	hasDigit := false
	for _, r := range text {
		if r >= '0' && r <= '9' {
			hasDigit = true
			break
		}
	}
	if hasDigit {
		return true
	}

	numberSignals := []string{
		"percent",
		"%",
		"billion",
		"million",
		"milyar",
		"milyon",
		"yüzde",
		"varil",
		"barrel",
	}

	for _, kw := range numberSignals {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func hasStrongSecondarySignal(text string) bool {
	secondarySignals := []string{
		"market impact",
		"financial stability",
		"supply shock",
		"energy shock",
		"global markets",
		"stock futures",
		"oil prices",
		"fuel prices",
		"market volatility",
		"arz şoku",
		"küresel piyasalar",
		"piyasa etkisi",
		"enerji krizi",
		"arz güvenliği",
		"tedarik riski",
	}

	for _, kw := range secondarySignals {
		if strings.Contains(text, kw) {
			return true
		}
	}

	return false
}
