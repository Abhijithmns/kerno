// Copyright 2026 Optiqor contributors
// SPDX-License-Identifier: Apache-2.0

package chaos

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// MemoryScenario allocates memory progressively over its run window,
// touching every page so the kernel actually commits the pages. Pairs
// with the oom_imminent rule (predicts an OOM kill from growth rate).
type MemoryScenario struct{}

func init() { Register(MemoryScenario{}) }

// Name implements Scenario.
func (MemoryScenario) Name() string { return "memory" }

// Description implements Scenario.
func (MemoryScenario) Description() string {
	return "Grow resident memory steadily, touching every page"
}

// PairedRule implements Scenario.
func (MemoryScenario) PairedRule() string { return "oom_imminent" }

// Run implements Scenario.
func (s MemoryScenario) Run(ctx context.Context, opts Options) error {
	targetMB := memoryMBFromIntensity(opts.Intensity)
	fmt.Fprintf(opts.Out, "    allocating up to %d MB over %s\n", targetMB, opts.Duration)

	chunkBytes := 1 << 20 // 1 MB
	chunks := make([][]byte, 0, targetMB)

	growInterval := opts.Duration / time.Duration(targetMB)
	if growInterval <= 0 {
		growInterval = time.Millisecond
	}
	ticker := time.NewTicker(growInterval)
	defer ticker.Stop()

	for len(chunks) < targetMB {
		select {
		case <-ctx.Done():
			runtime.KeepAlive(chunks)
			return nil
		case <-ticker.C:
			buf := make([]byte, chunkBytes)
			// Touch every page (4K) so the kernel commits real RSS,
			// not just virtual address space.
			for i := 0; i < chunkBytes; i += 4096 {
				buf[i] = 0xff
			}
			chunks = append(chunks, buf)
		}
	}

	fmt.Fprintf(opts.Out, "    held %d MB resident; idling until duration expires\n", len(chunks))
	<-ctx.Done()
	runtime.KeepAlive(chunks)
	return nil
}

func memoryMBFromIntensity(intensity Intensity) int {
	switch intensity {
	case IntensityLow:
		return 64
	case IntensityHigh:
		return 512
	default:
		return 200
	}
}
