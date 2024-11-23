package lib

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
)

// A "resource" must be json serializable/unserializable
type Resource interface {
	Create(ctx context.Context) error
	Delete(ctx context.Context) error

	// returns a unique identifier for the resource type
	Type() string
	// returns a string representation of a resource
	String() string
	// "other" is guaranteed to be a Resource with the same "Type()"
	Eq(other Resource) bool
}

func getRemovals(current, target []Resource) []Resource {
	var removals []Resource
current:
	for _, cur := range current {
		for _, tar := range target {
			if tar.Eq(cur) {
				continue current
			}
		}
		removals = append(removals, cur)
	}
	return removals
}

func getCreations(current, target []Resource) []Resource {
	var creations []Resource
current:
	for _, tar := range target {
		for _, cur := range current {
			if cur.Eq(tar) {
				continue current
			}
		}
		creations = append(creations, tar)
	}
	return creations
}

func reportTransformError(ctx context.Context, res Resource, err error) {
	slog.ErrorContext(
		ctx, "resource transformation failure",
		"resource", res.String(),
		"error", err,
	)
}

func Transform(ctx context.Context, current, target []Resource) {
	ctx = context.WithValue(ctx, "component", "transform")

	// create resources

	wg := sync.WaitGroup{}
	jobs := make(chan Resource)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for res := range jobs {
				err := res.Create(ctx)
				wg.Done()
				if err != nil {
					reportTransformError(ctx, res, fmt.Errorf("create resource: %w", err))
				} else {
					slog.InfoContext(ctx, "successfully created resource", "resource", res.String())
				}
			}
		}()
	}

	creations := getCreations(current, target)
	for _, res := range creations {
		wg.Add(1)
		jobs <- res
	}
	wg.Wait()
	close(jobs)

	// delete resources

	wg = sync.WaitGroup{}
	jobs = make(chan Resource)
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for res := range jobs {
				err := res.Delete(ctx)
				wg.Done()
				if err != nil {
					reportTransformError(ctx, res, fmt.Errorf("delete resource: %w", err))
				} else {
					slog.InfoContext(ctx, "successfully removed resource", "resource", res.String())
				}
			}
		}()
	}

	removals := getRemovals(current, target)
	for _, res := range removals {
		wg.Add(1)
		jobs <- res
	}
	wg.Wait()
	close(jobs)
}
