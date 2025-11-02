package storage

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/yourname/searchsvc/internal/normalizer"
)

type Repo struct{ db *sql.DB }
func NewRepo(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) UpsertContents(ctx context.Context, items []normalizer.StandardContent) error {
    if len(items) == 0 { return nil }
    tx, err := r.db.BeginTx(ctx, nil); if err != nil { return err }
    stmt := `
    INSERT INTO contents (
        external_id, provider, title, type, views, likes, reactions, reading_time_min,
        published_at, score_popularity, score_relevance, updated_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())
    ON CONFLICT (external_id, provider) DO UPDATE SET
        title=EXCLUDED.title,
        type=EXCLUDED.type,
        views=EXCLUDED.views,
        likes=EXCLUDED.likes,
        reactions=EXCLUDED.reactions,
        reading_time_min=EXCLUDED.reading_time_min,
        published_at=EXCLUDED.published_at,
        score_popularity=EXCLUDED.score_popularity,
        score_relevance=EXCLUDED.score_relevance,
        updated_at=now()
    `
    ps, err := tx.PrepareContext(ctx, stmt); if err != nil { _=tx.Rollback(); return err }
    defer ps.Close()
    for _, it := range items {
        if _, err := ps.ExecContext(ctx,
            it.ExternalID, it.Provider, it.Title, it.Type, it.Views, it.Likes, it.Reactions, it.ReadingTimeMin,
            it.PublishedAt, it.ScorePopularity, it.ScoreRelevance,
        ); err != nil { _=tx.Rollback(); return err }
    }
    return tx.Commit()
}

// RunMigrations executes all .sql files in dir (sorted by name)
func RunMigrations(ctx context.Context, db *sql.DB, dir string) error {
    ents, err := os.ReadDir(dir)
    if err != nil { return fmt.Errorf("read migrations dir: %w", err) }
    var files []string
    for _, e := range ents {
        if e.IsDir() { continue }
        name := e.Name()
        if strings.HasSuffix(name, ".sql") {
            files = append(files, filepath.Join(dir, name))
        }
    }
    sort.Strings(files)
    for _, fp := range files {
        b, err := os.ReadFile(fp); if err != nil { return fmt.Errorf("read %s: %w", fp, err) }
        if _, err := db.ExecContext(ctx, string(b)); err != nil { return fmt.Errorf("apply %s: %w", fp, err) }
    }
    return nil
}
