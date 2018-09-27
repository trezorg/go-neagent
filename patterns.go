package main

import (
	"context"
)

func orDone(ctx context.Context, input <-chan strResult) <-chan strResult {
	out := make(chan strResult)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return out
}

func bridge(ctx context.Context, inputs <-chan chan strResult) <-chan strResult {
	out := make(chan strResult)
	go func() {
		defer close(out)
		for {
			var stream <-chan strResult
			select {
			case maybe, ok := <-inputs:
				if !ok {
					return
				}
				stream = maybe
			case <-ctx.Done():
				return
			}
			for val := range orDone(ctx, stream) {
				select {
				case out <- val:
				case <-ctx.Done():
				}
			}
		}
	}()
	return out
}
