package proxy

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type CacheEntry struct {
	Body     []byte
	Response *http.Response
	CachedAt time.Time
}

type Proxy struct {
	Origin string
	Cache  map[string]*CacheEntry
	mu     sync.RWMutex
}

func NewProxy(origin string) *Proxy {
	return &Proxy{
		Origin: origin,
		Cache:  make(map[string]*CacheEntry),
		mu:     sync.RWMutex{},
	}
}

func (p *Proxy) ClearCache() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Cache = make(map[string]*CacheEntry)
	fmt.Printf("%d cache cleared at %s\n", len(p.Cache), time.Now().Format(time.RFC3339))
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cacheKey := r.Method + ":" + r.URL.String()

	p.mu.RLock()
	if entry, ok := p.Cache[cacheKey]; ok {
		fmt.Printf("CACHE HIT: %s\n", cacheKey)
		for k, v := range entry.Response.Header {
			w.Header()[k] = v
		}
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(entry.Response.StatusCode)
		w.Write(entry.Body)

		return
	}
	p.mu.RUnlock()

	res, err := http.Get(p.Origin + r.URL.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error forwarding request: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error forwarding request body: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	p.mu.Lock()
	p.Cache[cacheKey] = &CacheEntry{
		Body:     body,
		Response: res,
		CachedAt: time.Now(),
	}
	p.mu.Unlock()

	fmt.Printf("CACHE MISS: %s\n", cacheKey)
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(res.StatusCode)
	w.Write(body)
}
