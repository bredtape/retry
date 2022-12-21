package retry

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

func validateJitter(ratio float64) error {
	if ratio < 0 {
		return errors.New("jitter ratio must not be negative")
	}
	if ratio >= 0.5 {
		return errors.New("jitter ratio must be less than 0.5")
	}
	return nil
}

func withRandomJitter(ratio float64, d time.Duration) time.Duration {
	if ratio == 0 {
		return d
	}
	jitter := ratio * (rand.Float64() - 0.5)
	return scaleDuration(1+jitter, d)
}

func scaleDuration(factor float64, d time.Duration) time.Duration {
	return time.Duration(math.Round(float64(d.Nanoseconds())*factor)) * time.Nanosecond
}
