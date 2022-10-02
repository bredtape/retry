package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// retry with specified 'backoff' sequence and jitter
type Seq struct {
	xs     []time.Duration
	jitter float64
}

// new retry Sequence with jitter ratio [0, 0.5) and sequence.
// durations may be 0, but not negative
func NewSeq(jitterRatio float64, xs ...time.Duration) (Seq, error) {
	seq := Seq{
		xs:     xs,
		jitter: jitterRatio}

	if len(xs) == 0 {
		return seq, errors.New("empty sequence")
	}

	for _, x := range xs {
		if x < 0 {
			return seq, fmt.Errorf("duration must be positive")
		}
	}

	return seq, validateJitter(jitterRatio)
}

func (s Seq) Next(n int) time.Duration {
	if n < 0 {
		panic("n must not be negative")
	}

	lastIdx := len(s.xs) - 1
	if n > lastIdx {
		n = lastIdx
	}

	return withRandomJitter(s.jitter, s.xs[n])
}

func (s Seq) Try(ctx context.Context, f func() error) error {
	n := 0
	for {
		err := f()
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.Next(n)):
			n++
		}
	}
}

func (s Seq) MaxDuration() time.Duration {
	max := time.Duration(0)
	for _, x := range s.xs {
		if x > max {
			max = x
		}
	}

	return scaleDuration(1+s.jitter, max)
}
