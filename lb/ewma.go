package main

import "math"

// EWMA helper (stateless alpha calculation)
func AlphaFromDtTau(dt, tau float64) float64 {
	if tau <= 0 { return 1.0 }
	return 1 - math.Exp(-dt/tau)
}
