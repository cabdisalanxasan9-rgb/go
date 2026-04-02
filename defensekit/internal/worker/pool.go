package worker

import (
	"context"
	"sync"
	"time"
)

type Job[T any, R any] func(context.Context, T) R

func Run[T any, R any](ctx context.Context, inputs []T, workers, rate int, timeout time.Duration, job Job[T, R]) []R {
	if workers < 1 {
		workers = 1
	}
	if rate < 1 {
		rate = 1
	}

	jobs := make(chan T)
	results := make(chan R, len(inputs))
	var wg sync.WaitGroup

	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for in := range jobs {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}

				jobCtx, cancel := context.WithTimeout(ctx, timeout)
				res := job(jobCtx, in)
				cancel()
				results <- res
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, in := range inputs {
			select {
			case <-ctx.Done():
				return
			case jobs <- in:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]R, 0, len(inputs))
	for r := range results {
		out = append(out, r)
	}
	return out
}
