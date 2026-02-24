package config

type Config struct {
	RSSUrls []string // Takip edilecek RSS adresleri
}

// LoadConfig şimdilik elle girilen ayarları döner.
// İleride burası .env dosyasından okuyacak şekilde güncellenicek.
func LoadConfig() *Config {
	return &Config{
		RSSUrls: []string{
			"https://feeds.feedburner.com/TechCrunch/", // Teknoloji Haberleri
			"https://news.ycombinator.com/rss",         // Hacker News
			"https://openai.com/blog/rss.xml",          // AI Haberleri (Proje için önemli)
			"https://feeds.bbci.co.uk/news/technology/rss.xml",
		},
	}
}
