package utils

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(count float64, seconds int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(count/float64(seconds)), int(count)),
	}
}

func (r *RateLimiter) Limit(ctx context.Context, next func() error) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		err := r.limiter.Wait(ctx)
		if err == nil {
			err = next()
		}
		ch <- err
	}()
	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("request did not complete in time due to rate limit")
	case err := <-ch:
		return err
	}
}
