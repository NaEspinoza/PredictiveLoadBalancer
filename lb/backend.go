package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

// Backend representa un servidor destino gestionado por el LB
type Backend struct {
	URL          *url.URL
	Alive        int32   // atomic: 1 true, 0 false
	EWMALatency  float64 // ms (protected by ewmaMux)
	ewmaMux      sync.RWMutex
	Connections  int64 // atomic
	ErrorCount   int64 // atomic
	ReverseProxy *httputil.ReverseProxy
	BaselineMs   float64
	MaxConn      int64
}

// NewBackend crea un backend desde una URL
func NewBackend(rawurl string) (*Backend, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	b := &Backend{URL: u, ReverseProxy: proxy, Alive: 1, BaselineMs: 50, MaxConn: 100, EWMALatency: 50}
	return b, nil
}

func (b *Backend) IncConn()        { atomic.AddInt64(&b.Connections, 1) }
func (b *Backend) DecConn()        { atomic.AddInt64(&b.Connections, -1) }
func (b *Backend) GetConn() int64  { return atomic.LoadInt64(&b.Connections) }
func (b *Backend) AddError()       { atomic.AddInt64(&b.ErrorCount, 1) }
func (b *Backend) ResetErrors()    { atomic.StoreInt64(&b.ErrorCount, 0) }
func (b *Backend) IsAlive() bool   { return atomic.LoadInt32(&b.Alive) == 1 }
func (b *Backend) SetAlive(v bool) {
	if v {
		atomic.StoreInt32(&b.Alive, 1)
	} else {
		atomic.StoreInt32(&b.Alive, 0)
	}
}

func (b *Backend) UpdateEWMA(latencyMs, alpha float64) {
	b.ewmaMux.Lock()
	defer b.ewmaMux.Unlock()
	b.EWMALatency = alpha*latencyMs + (1-alpha)*b.EWMALatency
}

func (b *Backend) GetEWMA() float64 {
	b.ewmaMux.RLock()
	defer b.ewmaMux.RUnlock()
	return b.EWMALatency
}

func (b *Backend) GetScore(wLat, wConn, errPenaltyMs float64) float64 {
	lat := b.GetEWMA()
	conn := float64(b.GetConn())
	latScore := lat / b.BaselineMs
	connScore := conn / float64(b.MaxConn)
	errs := float64(atomic.LoadInt64(&b.ErrorCount))
	pen := (errPenaltyMs / b.BaselineMs) * errCap(errs)
	return wLat*latScore + wConn*connScore + pen
}

func errCap(e float64) float64 {
	if e > 10 {
		return 10
	}
	return e
}
