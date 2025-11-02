package normalizer

import (
	"time"

	prov "github.com/yourname/searchsvc/internal/provider"
)

type StandardContent struct {
	ExternalID      string
	Provider        string
	Title           string
	Type            string // video|text
	Views           int
	Likes           int
	Reactions       int
	ReadingTimeMin  int
	PublishedAt     time.Time
	ScorePopularity float64
	ScoreRelevance  float64
}

// DİKKAT: Artık tek dilim alıyor: []prov.RawItem
func Normalize(items []prov.RawItem) []StandardContent {
	out := make([]StandardContent, 0, len(items))
	for _, r := range items {
		out = append(out, StandardContent{
			ExternalID:     r.ExternalID,
			Provider:       r.Provider,
			Title:          r.Title,
			Type:           r.Type,
			Views:          r.Views,
			Likes:          r.Likes,
			Reactions:      r.Reactions,
			ReadingTimeMin: r.ReadMin,
			PublishedAt:    r.Published,
		})
	}
	return out
}
