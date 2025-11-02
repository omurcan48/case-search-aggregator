package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/yourname/searchsvc/internal/services"
)

type SearchHandler struct{ svc *services.SearchService }

func NewSearchRouter(svc *services.SearchService) http.Handler {
    h := &SearchHandler{svc: svc}
    r := chi.NewRouter()
    r.Get("/", h.search)
    return r
}

func (h *SearchHandler) search(w http.ResponseWriter, r *http.Request) {
    q := strings.TrimSpace(r.URL.Query().Get("query"))
    if q == "" { writeErr(w, http.StatusBadRequest, "query is required"); return }
    typ := r.URL.Query().Get("type")
    if typ != "" && typ != "video" && typ != "text" { writeErr(w, 400, "type must be video|text"); return }
    sort := r.URL.Query().Get("sort")
    if sort != "" && sort != "popularity" && sort != "relevance" { writeErr(w, 400, "sort must be popularity|relevance"); return }
    page := atoiDef(r.URL.Query().Get("page"), 1)
    size := atoiDef(r.URL.Query().Get("size"), 20)
    if page < 1 || size < 1 || size > 100 { writeErr(w, 400, "page>=1, 1<=size<=100"); return }

    resp, err := h.svc.Search(r.Context(), services.SearchParams{
        Query: q, Type: typ, Sort: sort, Page: page, Size: size,
    })
    if err != nil { writeErr(w, 500, err.Error()); return }
    writeJSON(w, 200, resp)
}

func atoiDef(s string, d int) int { if s=="" {return d}; n,err:=strconv.Atoi(s); if err!=nil {return d}; return n }

func writeErr(w http.ResponseWriter, code int, msg string) { writeJSON(w, code, map[string]string{"error": msg}) }
func writeJSON(w http.ResponseWriter, code int, v any) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(code); _=json.NewEncoder(w).Encode(v) }
