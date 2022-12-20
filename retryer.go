package retry

import (
	"context"
	"time"
)

type Retryer interface {
	// calculate the duration for the n'th index (starts with 0)
	Next(n int) time.Duration

	// keep trying the function f until it returns nil or the context has expired
	// this operation will block
	Try(context.Context, func() error) error

	// maximum duration that can occur between attempts
	MaxDuration() time.Duration
}

func Must(r Retryer, err error) Retryer {
	if err != nil {
		panic(err)
	}
	return r
}
