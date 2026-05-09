// Copyright 2026 Optiqor contributors
// SPDX-License-Identifier: Apache-2.0

package chaos

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
)

// CPUScenario saturates the host CPU by running tight loops on multiple
// goroutines. Pairs with the scheduler_contention rule (run-queue delay
// climbs once N > NumCPU).
type CPUScenario struct{}

func init() { Register(CPUScenario{}) }

// Name implements Scenario.
func (CPUScenario) Name() string { return "cpu" }

// Description implements Scenario.
func (CPUScenario) Description() string {
	return "Pin N goroutines on tight CPU loops to drive scheduler contention"
}

// PairedRule implements Scenario.
func (CPUScenario) PairedRule() string { return "scheduler_contention" }

// Run implements Scenario.
func (s CPUScenario) Run(ctx context.Context, opts Options) error {
	workers := workersFromIntensity(opts.Intensity, runtime.NumCPU())
	fmt.Fprintf(opts.Out, "    spawning %d CPU-saturation workers\n", workers)

	// sink is updated atomically across all workers so the compiler
	// can't prove the loop body is dead.
	var sink atomic.Uint64

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(seed int64) {
			defer wg.Done()
			// math/rand is fine here — we only need pseudo-random
			// floats to keep the optimizer from removing the loop.
			r := rand.New(rand.NewSource(seed)) //nolint:gosec
			for ctx.Err() == nil {
				var local float64
				for k := 0; k < 100_000 && ctx.Err() == nil; k++ {
					local += math.Sqrt(r.Float64())
				}
				sink.Add(uint64(local))
			}
		}(int64(i))
	}
	wg.Wait()
	_ = sink.Load() // observe the running total to keep it live
	return nil
}
