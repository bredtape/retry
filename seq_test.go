package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

func ExampleSeq_Next() {
	// new sequence with no jitter (to be able to match output in this example)
	r := Must(NewSeq(0, 100*time.Millisecond, 1*time.Second, 2*time.Second))

	fmt.Println(r.Next(0).String())
	fmt.Println(r.Next(1).String())
	fmt.Println(r.Next(2).String())
	fmt.Println(r.Next(3).String())

	// Output:
	// 100ms
	// 1s
	// 2s
	// 2s
}

func TestSeqNextNoJitter(t *testing.T) {
	r, err := NewSeq(0, 100*time.Millisecond, 20*time.Second, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expected := []float64{0.1, 20, 10, 10}

	for i := range expected {
		d := r.Next(i)

		diff := math.Abs(d.Seconds() - expected[i])
		if diff > 1e-3 {
			t.Errorf("index %d, backoff duration expected %f, but was %f (diff %f)", i, expected[i], d.Seconds(), diff)
		}
	}
}

func TestSeqNextWithJitter(t *testing.T) {
	ratio := 0.2

	r, err := NewSeq(ratio, 100*time.Millisecond, 20*time.Second, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expected := []float64{0.1, 20, 10, 10}

	for i := range expected {
		d := r.Next(i)

		diff := math.Abs(d.Seconds() - expected[i])
		if diff > expected[i]*ratio*0.5 {
			t.Errorf("index %d, backoff duration expected %f, but was %f (diff %f)", i, expected[i], d.Seconds(), diff)
		}
	}
}

func TestSeqMaxDurationWithNoJitter(t *testing.T) {
	r, err := NewSeq(0, 100*time.Millisecond, 20*time.Second, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if r.MaxDuration() != 20*time.Second {
		t.Errorf("expected max duration to be 20sec, was %s", r.MaxDuration().String())
	}
}

func TestSeqTry(t *testing.T) {

	r, err := NewSeq(0.1, 1*time.Millisecond, 10*time.Millisecond, 5*time.Millisecond, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	N := 3
	n := 0

	f := func() error {
		n++
		if n == N {
			return nil
		}
		return errors.New("are we there yet?")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = r.Try(ctx, f)

	if n != N {
		t.Errorf("the function should have been run %d times, but was %d", N, n)
	}
	if err != nil {
		t.Error(err)
	}
}
