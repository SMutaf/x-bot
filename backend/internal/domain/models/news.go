package models

import (
	"fmt"
	"strings"
	"time"
)

type NewsCategory string

const (
	CategoryBreaking NewsCategory = "BREAKING"
	CategoryTech     NewsCategory = "TECH"
	CategoryGeneral  NewsCategory = "GENERAL"
	CategoryEconomy  NewsCategory = "ECONOMY"
	CategorySports   NewsCategory = "SPORTS"
	CategoryScience  NewsCategory = "SCIENCE"
)

type RawNewsItem struct {
	ID          string
	Title       string
	Description string
	Link        string
	Source      string
	Category    NewsCategory
	PublishedAt time.Time
	FetchedAt   time.Time
}

func (r RawNewsItem) EffectiveTime() time.Time {
	if !r.PublishedAt.IsZero() {
		return r.PublishedAt
	}
	return r.FetchedAt
}

func (r RawNewsItem) BuildID() string {
	base := strings.TrimSpace(strings.ToLower(r.Source + "|" + r.Title + "|" + r.Link))
	if base == "" {
		return fmt.Sprintf("news-%d", time.Now().UnixNano())
	}
	return base
}
