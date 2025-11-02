package services

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/yourname/searchsvc/internal/cache"
	"github.com/yourname/searchsvc/internal/normalizer"
	"github.com/yourname/searchsvc/internal/provider"
	"github.com/yourname/searchsvc/internal/scoring"
	"github.com/yourname/searchsvc/internal/storage"
)

type SearchParams struct {
	Query string
	Type  string // "", "video", "text"
	Sort  string // "", "popularity", "relevance"
	Page  int
	Size  int
}

type ItemDTO struct {
	ExternalID  string  `json:"external_id"`
	Provider    string  `json:"provider"`
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	Score       float64 `json:"score"`
	PublishedAt string  `json:"published_at"`
}

type SearchResponse struct {
	Items         []ItemDTO `json:"items"`
	Page          int       `json:"page"`
	Size          int       `json:"size"`
	TotalEstimate int       `json:"total_estimate"`
}

type SearchService struct {
	providers []provider.Provider
	repo      *storage.Repo
	cache     *cache.TTLCache
}

func NewSearchService(ps []provider.Provider, repo *storage.Repo, c *cache.TTLCache) *SearchService {
	return &SearchService{providers: ps, repo: repo, cache: c}
}

func (s *SearchService) Search(ctx context.Context, p SearchParams) (*SearchResponse, error) {
	// Cache key (query'yi normalize ederek)
	key := strings.Join([]string{
		strings.ToLower(strings.TrimSpace(p.Query)),
		p.Type, p.Sort, itoa(p.Page), itoa(p.Size),
	}, "|")
	if v, ok := s.cache.Get(key); ok {
		if resp, ok2 := v.(*SearchResponse); ok2 {
			return resp, nil
		}
	}

	// 1) Provider'lardan çek
	raws, _ := provider.FetchAll(ctx, s.providers, p.Query)

	// 2) Normalize
	var contents []normalizer.StandardContent
	for _, rr := range raws {
		contents = append(contents, normalizer.Normalize(rr)...)
	}

	// 3) Query filtresi (title içinde tüm kelimeler geçmeli, case-insensitive)
	if q := strings.ToLower(strings.TrimSpace(p.Query)); q != "" {
		tokens := strings.Fields(q)
		dst := contents[:0]
	OUTER:
		for _, c := range contents {
			title := strings.ToLower(c.Title)
			for _, t := range tokens {
				if !strings.Contains(title, t) {
					continue OUTER
				}
			}
			dst = append(dst, c)
		}
		contents = dst
	}

	// 4) Tip filtresi
	if p.Type == "video" || p.Type == "text" {
		filtered := contents[:0]
		for _, c := range contents {
			if c.Type == p.Type {
				filtered = append(filtered, c)
			}
		}
		contents = filtered
	}

	// 5) Skor hesapla
	for i := range contents {
		pop, rel := scoring.Score(contents[i])
		contents[i].ScorePopularity = pop
		contents[i].ScoreRelevance = rel
	}

	// 6) Sıralama
	switch p.Sort {
	case "relevance":
		sort.Slice(contents, func(i, j int) bool { return contents[i].ScoreRelevance > contents[j].ScoreRelevance })
	default:
		sort.Slice(contents, func(i, j int) bool { return contents[i].ScorePopularity > contents[j].ScorePopularity })
	}

	// 7) Sayfalama
	total := len(contents)
	start := (p.Page - 1) * p.Size
	if start >= total {
		resp := &SearchResponse{Items: []ItemDTO{}, Page: p.Page, Size: p.Size, TotalEstimate: total}
		s.cache.Set(key, resp)
		return resp, nil
	}
	end := start + p.Size
	if end > total {
		end = total
	}
	pageItems := contents[start:end]

	// 8) DB'ye UPSERT
	if err := s.repo.UpsertContents(ctx, pageItems); err != nil {
		return nil, err
	}

	// 9) DTO
	items := make([]ItemDTO, 0, len(pageItems))
	for _, c := range pageItems {
		score := c.ScorePopularity
		if p.Sort == "relevance" {
			score = c.ScoreRelevance
		}
		items = append(items, ItemDTO{
			ExternalID:  c.ExternalID,
			Provider:    c.Provider,
			Title:       c.Title,
			Type:        c.Type,
			Score:       score,
			PublishedAt: c.PublishedAt.Format(time.RFC3339),
		})
	}

	resp := &SearchResponse{Items: items, Page: p.Page, Size: p.Size, TotalEstimate: total}
	s.cache.Set(key, resp)
	return resp, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(b[i:])
}
