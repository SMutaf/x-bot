package dedup

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Deduplicator struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewDeduplicator(addr string) *Deduplicator {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &Deduplicator{
		Client: rdb,
		Ctx:    context.Background(),
	}
}

func (d *Deduplicator) slugify(text string) string {
	re := regexp.MustCompile("[^a-z0-9]+")
	return strings.Trim(re.ReplaceAllString(strings.ToLower(text), "-"), "-")
}

func (d *Deduplicator) IsDuplicate(url string) bool {
	exists, err := d.Client.Exists(d.Ctx, url).Result()
	if err != nil {
		fmt.Printf("Redis Hatası: %v\n", err)
		return false
	}

	if exists > 0 {
		return true
	}

	err = d.Client.Set(d.Ctx, url, "seen", 7*24*time.Hour).Err()
	if err != nil {
		fmt.Printf("Redis Kayıt Hatası: %v\n", err)
	}

	return false
}

func (d *Deduplicator) IsTitleDuplicate(title string) bool {
	slug := "title:" + d.slugify(title)
	exists, _ := d.Client.Exists(d.Ctx, slug).Result()
	if exists > 0 {
		return true
	}
	// Başlığı 24 saat boyunca saklar
	d.Client.Set(d.Ctx, slug, "seen", 24*time.Hour)
	return false
}
