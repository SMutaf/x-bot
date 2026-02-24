package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 1. Örnek bir RSS kaynağı (Hacker News)
	rssURL := "https://news.ycombinator.com/rss"
	
	fmt.Println("Haberler çekiliyor: " + rssURL)

	// 2. HTTP İsteği Atmak (Timeout ekledik ki sonsuza kadar beklemesin)
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(rssURL)
	if err != nil {
		fmt.Println("Hata oluştu:", err)
		return
	}
	defer resp.Body.Close()

	// 3. Cevabı Okumak
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Veri okunamadı:", err)
		return
	}

	// 4. Ekrana Basmak (İlk 500 karakteri basalım, ekran dolmasın)
	fmt.Println("--- Gelen Ham Veri (İlk 500 Karakter) ---")
	fmt.Println(string(body)[:500]) 
	fmt.Println("...")
	fmt.Println("--- Başarılı! ---")
}