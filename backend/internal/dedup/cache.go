package dedup

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Deduplicator struct {
	Client *redis.Client
	Ctx    context.Context
}

// NewDeduplicator yeni bir Redis bağlantısı oluşturur
func NewDeduplicator(addr string) *Deduplicator {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // Örn: "localhost:6379"
		Password: "",   // Şifre yoksa boş bırak
		DB:       0,    // Varsayılan veritabanı
	})

	return &Deduplicator{
		Client: rdb,
		Ctx:    context.Background(),
	}
}

// IsDuplicate linki kontrol eder.
// Eğer link daha önce VARSA -> true döner (İşleme!)
// Eğer link YOKSA -> Redis'e kaydeder ve false döner (İşle!)
func (d *Deduplicator) IsDuplicate(url string) bool {
	// 1. Önce var mı diye soruyoruz
	exists, err := d.Client.Exists(d.Ctx, url).Result()
	if err != nil {
		fmt.Printf("⚠️ Redis Hatası: %v\n", err)
		return false // Hata varsa işlemeye devam etsin, akışı bozmayalım
	}

	if exists > 0 {
		return true // Evet, bu linki daha önce kaydetmişiz
	}

	// 2. Yoksa kaydediyoruz (7 gün sonra silinsin diye TTL ekledik)
	err = d.Client.Set(d.Ctx, url, "seen", 7*24*time.Hour).Err()
	if err != nil {
		fmt.Printf("⚠️ Redis Kayıt Hatası: %v\n", err)
	}

	return false // Hayır, bu yeni bir link
}
