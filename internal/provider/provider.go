package provider

import (
    "context"
    "encoding/json"
    "encoding/xml"
    "errors"
    "io"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"

    "github.com/yourname/searchsvc/internal/rate"
)

type RawItem struct {
    ExternalID string    `json:"id" xml:"id"`
    Title      string    `json:"title" xml:"title"`
    Type       string    `json:"type" xml:"type"`
    Views      int       `json:"views" xml:"views"`
    Likes      int       `json:"likes" xml:"likes"`
    Reactions  int       `json:"reactions" xml:"reactions"`
    ReadMin    int       `json:"reading_time_min" xml:"reading_time_min"`
    Published  time.Time `json:"published_at" xml:"published_at"`
    Provider   string    `json:"-" xml:"-"`
}

type Provider interface {
    Name() string
    Fetch(ctx context.Context, query string) ([]RawItem, error)
}

func httpGet(ctx context.Context, endpoint string, query string) ([]byte, error) {
    if strings.HasPrefix(endpoint, "file://") {
        path := strings.TrimPrefix(endpoint, "file://")
        return os.ReadFile(path)
    }
    u, err := url.Parse(endpoint)
    if err != nil { return nil, err }
    q := u.Query()
    if query != "" { q.Set("q", query) }
    u.RawQuery = q.Encode()

    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    client := &http.Client{ Timeout: 5 * time.Second }
    resp, err := client.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, errors.New("provider status " + resp.Status)
    }
    return io.ReadAll(resp.Body)
}

type JSONProvider struct {
    name    string
    baseURL string
    limiter *rate.Registry
}

func NewJSONProvider(name, baseURL string, limiter *rate.Registry) *JSONProvider {
    return &JSONProvider{name: name, baseURL: baseURL, limiter: limiter}
}
func (p *JSONProvider) Name() string { return p.name }
func (p *JSONProvider) Fetch(ctx context.Context, query string) ([]RawItem, error) {
    p.limiter.Take(ctx, p.name)
    b, err := httpGet(ctx, p.baseURL, query)
    if err != nil { return nil, err }
    var arr []RawItem
    if err := json.Unmarshal(b, &arr); err != nil { return nil, err }
    for i := range arr { arr[i].Provider = p.name }
    return arr, nil
}

type xmlItems struct { Items []RawItem `xml:"item"` }

type XMLProvider struct {
    name    string
    baseURL string
    limiter *rate.Registry
}

func NewXMLProvider(name, baseURL string, limiter *rate.Registry) *XMLProvider {
    return &XMLProvider{name: name, baseURL: baseURL, limiter: limiter}
}
func (p *XMLProvider) Name() string { return p.name }
func (p *XMLProvider) Fetch(ctx context.Context, query string) ([]RawItem, error) {
    p.limiter.Take(ctx, p.name)
    b, err := httpGet(ctx, p.baseURL, query)
    if err != nil { return nil, err }
    var wrap xmlItems
    if err := xml.Unmarshal(b, &wrap); err != nil { return nil, err }
    for i := range wrap.Items { wrap.Items[i].Provider = p.name }
    return wrap.Items, nil
}

// FetchAll fetches from all providers in parallel; errors ignored for partial success.
func FetchAll(ctx context.Context, providers []Provider, query string) ([][]RawItem, error) {
    if len(providers) == 0 { return nil, errors.New("no providers") }
    type res struct { items []RawItem }
    ch := make(chan res, len(providers))
    for _, p := range providers {
        p := p
        go func() {
            items, _ := p.Fetch(ctx, query)
            ch <- res{items: items}
        }()
    }
    out := make([][]RawItem, 0, len(providers))
    for i := 0; i < len(providers); i++ {
        r := <-ch
        out = append(out, r.items)
    }
    return out, nil
}
