package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// LoadBalancer contiene la lógica P2C + EWMA
type LoadBalancer struct {
	backends     []*Backend
	alpha        float64
	wLat         float64
	wConn        float64
	errPenaltyMs float64
	rng          *rand.Rand
}

func NewLoadBalancer(backends []*Backend, alpha, wLat, wConn, errPenaltyMs float64) *LoadBalancer {
	return &LoadBalancer{
		backends:     backends,
		alpha:        alpha,
		wLat:         wLat,
		wConn:        wConn,
		errPenaltyMs: errPenaltyMs,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (lb *LoadBalancer) Select() *Backend {
	n := len(lb.backends)
	if n == 0 {
		return nil
	}
	i1 := lb.rng.Intn(n)
	i2 := lb.rng.Intn(n)
	if i1 == i2 {
		i2 = (i1 + 1) % n
	}
	b1 := lb.backends[i1]
	b2 := lb.backends[i2]

	// prefer alive
	if b1.IsAlive() && !b2.IsAlive() {
		return b1
	}
	if b2.IsAlive() && !b1.IsAlive() {
		return b2
	}

	s1 := b1.GetScore(lb.wLat, lb.wConn, lb.errPenaltyMs)
	s2 := b2.GetScore(lb.wLat, lb.wConn, lb.errPenaltyMs)
	if s1 < s2 {
		return b1
	}
	return b2
}

func (lb *LoadBalancer) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	backend := lb.Select()
	if backend == nil {
		http.Error(w, "no backends", http.StatusServiceUnavailable)
		return
	}
	metricRequests.WithLabelValues(backend.URL.Host).Inc()

	backend.IncConn()
	defer backend.DecConn()

	start := time.Now().UnixNano()
	r.Header.Set("X-Proxy-Start", strconv.FormatInt(start, 10))
	r.Header.Set("X-Proxy-Backend", backend.URL.Host)

	proxy := backend.ReverseProxy
	proxy.Transport = http.DefaultTransport

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		metricErrors.WithLabelValues(backend.URL.Host).Inc()
		backend.AddError()
		backend.SetAlive(false)
		http.Error(rw, "backend error", http.StatusBadGateway)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		startStr := resp.Request.Header.Get("X-Proxy-Start")
		if startStr != "" {
			if ts, err := strconv.ParseInt(startStr, 10, 64); err == nil {
				latMs := float64(time.Now().UnixNano()-ts) / 1e6
				backend.UpdateEWMA(latMs, lb.alpha)
				metricLatency.WithLabelValues(backend.URL.Host).Observe(latMs)
			}
		}
		if resp.StatusCode >= 500 {
			backend.AddError()
		}
		resp.Header.Del("X-Proxy-Start")
		resp.Header.Del("X-Proxy-Backend")
		return nil
	}

	proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) StartHealthChecks(interval time.Duration, path string) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			for _, b := range lb.backends {
				go func(backend *Backend) {
					u := *backend.URL
					u.Path = path
					client := http.Client{Timeout: 2 * time.Second}
					resp, err := client.Get(u.String())
					if err != nil || resp.StatusCode >= 500 {
						backend.SetAlive(false)
						backend.AddError()
					} else {
						backend.SetAlive(true)
						backend.ResetErrors()
					}
					if resp != nil {
						resp.Body.Close()
					}
				}(b)
			}
		}
	}()
}
