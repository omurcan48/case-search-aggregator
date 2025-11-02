package scoring

import "github.com/yourname/searchsvc/internal/normalizer"

import "time"

func Score(c normalizer.StandardContent) (float64, float64) {
    base := 0.0
    if c.Type == "video" {
        base = float64(c.Views)/1000.0 + float64(c.Likes)/100.0
    } else {
        base = float64(c.Reactions)/50.0 + float64(c.ReadingTimeMin)
    }
    typeCoeff := 1.0
    if c.Type == "video" { typeCoeff = 1.5 }
    recency := recencyBoost(c.PublishedAt)
    interact := 0.0
    if c.Type == "video" {
        if c.Views > 0 { interact = (float64(c.Likes)/float64(c.Views))*10.0 }
    } else if c.ReadingTimeMin > 0 {
        interact = (float64(c.Reactions)/float64(c.ReadingTimeMin))*5.0
    }
    popularity := (base * typeCoeff) + interact
    relevance  := (base * typeCoeff) + recency
    return popularity, relevance
}

func recencyBoost(t time.Time) float64 {
    if t.IsZero() { return 0 }
    d := time.Since(t)
    switch {
    case d <= 7*24*time.Hour: return 5
    case d <= 30*24*time.Hour: return 3
    case d <= 90*24*time.Hour: return 1
    default: return 0
    }
}
