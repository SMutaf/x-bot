package models

import "time"

type NewsCategory string

const (
	CategoryBreaking NewsCategory = "BREAKING"
	CategoryTech     NewsCategory = "TECH"
	CategoryGeneral  NewsCategory = "GENERAL"
	CategoryEconomy  NewsCategory = "ECONOMY"
	CategorySports   NewsCategory = "SPORTS"
	CategoryScience  NewsCategory = "SCIENCE"
)

type NewsItem struct {
	Title        string
	Description  string
	Link         string
	Source       string
	Category     NewsCategory
	PublishedAt  time.Time
	Score        int
	ClusterCount int
	ClusterKey   string
}
