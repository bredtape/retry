package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"
)

func ExampleExp_Next() {
	// new sequence with no jitter (to be able to match output in this example)
	r := Must(NewExp(0, 300*time.Millisecond, 1*time.Second))

	fmt.Println(r.Next(0).String())
	fmt.Println(r.Next(1).String())
	fmt.Println(r.Next(2).String())
	fmt.Println(r.Next(3).String())

	// Output:
	// 300ms
	// 600ms
	// 1s
}

func TestExpNoJitter(t *testing.T) {

	r, err := NewExp(0, 100*time.Millisecond, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expected := []float64{0.1, 0.2, 0.4, 0.8, 1.6, 3.2, 5, 5, 5}

	for i := range expected {
		d := r.Next(i)

		diff := math.Abs(d.Seconds() - expected[i])
		if diff > 1e-3 {
			t.Errorf("index %d, backoff duratione expected %f, but was %f (diff %f)", i, expected[i], d.Seconds(), diff)
		}
	}
}

func TestExpWithJitter(t *testing.T) {

	ratio := 0.2
	r, err := NewExp(ratio, 100*time.Millisecond, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	expected := []float64{0.1, 0.2, 0.4, 0.8, 1.6, 3.2, 5, 5, 5}

	for i := range expected {
		d := r.Next(i)

		diff := math.Abs(d.Seconds() - expected[i])
		if diff > expected[i]*ratio*0.5 {
			t.Errorf("index %d, backoff duratione expected %f, but was %f (diff %f)", i, expected[i], d.Seconds(), diff)
		}
	}
}

func TestExpTry(t *testing.T) {

	r, err := NewExp(0.1, 1*time.Millisecond, 2*time.Second)
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
