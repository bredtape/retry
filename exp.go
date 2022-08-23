package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// exponential backoff with jitter and max duration
// the duration doubles after each retry
type Exp struct {
	// step duration in nano seconds
	step time.Duration

	// max duration in nano seconds
	max time.Duration

	// jitter ratio [0, 0.5)
	jitter float64
}

func NewExp(jitterRatio float64, step, max time.Duration) (Exp, error) {
	r := Exp{
		step:   step,
		max:    max,
		jitter: jitterRatio}

	if r.step <= 0 {
		return r, errors.New("step must be greater than 0")
	}

	if r.step >= r.max {
		return r, errors.New("step must be less than max")
	}

	return r, validateJitter(r.jitter)
}

func (r Exp) Next(n int) time.Duration {
	if n < 0 {
		panic("n must not be negative")
	}

	n_max := math.Log2(r.max.Seconds() / r.step.Seconds())
	if float64(n) > n_max {
		return withRandomJitter(r.jitter, r.max)
	}

	factor := math.Pow(2, float64(n))
	return withRandomJitter(r.jitter, scaleDuration(factor, r.step))
}

func (r Exp) Try(ctx context.Context, f func() error) error {
	n := 0

	for {
		err := f()
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.Next(n)):
			n++
		}
	}
}

func (r Exp) MaxDuration() time.Duration {
	return scaleDuration(1+r.jitter, r.max)
}
