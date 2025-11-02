package rate

import (
    "context"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

type Registry struct {
    mu   sync.Mutex
    m    map[string]*rate.Limiter
    perM int
}

func NewRegistry(perMinute int) *Registry {
    if perMinute <= 0 { perMinute = 60 }
    return &Registry{m: make(map[string]*rate.Limiter), perM: perMinute}
}

func (r *Registry) Take(ctx context.Context, name string) {
    r.mu.Lock()
    lim, ok := r.m[name]
    if !ok {
        lim = rate.NewLimiter(rate.Every(time.Minute/time.Duration(r.perM)), r.perM)
        r.m[name] = lim
    }
    r.mu.Unlock()
    _ = lim.Wait(ctx)
}
