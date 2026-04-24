package stream

import "sync"

type PublishedItem struct {
	Time         string `json:"time"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Hook         string `json:"hook"`
	Summary      string `json:"summary"`
	Importance   string `json:"importance"`
	Sentiment    string `json:"sentiment"`
	Category     string `json:"category"`
	Source       string `json:"source"`
	Link         string `json:"link"`
	Virality     int    `json:"virality"`
	ClusterCount int    `json:"clusterCount"`
}

type hub struct {
	mu          sync.RWMutex
	subscribers map[chan PublishedItem]struct{}
}

var publishedHub = &hub{
	subscribers: make(map[chan PublishedItem]struct{}),
}

func SubscribePublished() chan PublishedItem {
	ch := make(chan PublishedItem, 32)

	publishedHub.mu.Lock()
	publishedHub.subscribers[ch] = struct{}{}
	publishedHub.mu.Unlock()

	return ch
}

func UnsubscribePublished(ch chan PublishedItem) {
	publishedHub.mu.Lock()
	if _, ok := publishedHub.subscribers[ch]; ok {
		delete(publishedHub.subscribers, ch)
		close(ch)
	}
	publishedHub.mu.Unlock()
}

func PublishPublished(item PublishedItem) {
	publishedHub.mu.RLock()
	defer publishedHub.mu.RUnlock()

	for ch := range publishedHub.subscribers {
		select {
		case ch <- item:
		default:
		}
	}
}
