package main

import (
	"math"
	"testing"
)

func TestAlphaFromDtTau(t *testing.T) {
	dt := 0.1
	tau := 10.0
	alpha := AlphaFromDtTau(dt, tau)
	if alpha <= 0 || alpha >= 1 {
		t.Fatalf("alpha out of range: %f", alpha)
	}
}

func TestEWMAConvergence(t *testing.T) {
	alpha := AlphaFromDtTau(0.1, 1.0)
	b := &Backend{EWMALatency: 100}
	for i := 0; i < 1000; i++ {
		b.UpdateEWMA(10, alpha)
	}
	if math.Abs(b.GetEWMA()-10) > 1.0 {
		t.Fatalf("EWMA did not converge close to 10, got %f", b.GetEWMA())
	}
}
