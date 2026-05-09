// Copyright 2026 Optiqor contributors
// SPDX-License-Identifier: Apache-2.0

package chaos

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

// DiskScenario writes random data and fsyncs in a tight loop, driving
// disk write latency up. Pairs with the disk_io_bottleneck rule.
//
// To avoid filling the disk, the file is truncated periodically — only
// a small (~16 MB) working set is ever on disk.
type DiskScenario struct{}

func init() { Register(DiskScenario{}) }

// Name implements Scenario.
func (DiskScenario) Name() string { return "disk-sat" }

// Description implements Scenario.
func (DiskScenario) Description() string {
	return "Write+fsync in a tight loop to saturate block I/O latency"
}

// PairedRule implements Scenario.
func (DiskScenario) PairedRule() string { return "disk_io_bottleneck" }

// Run implements Scenario.
func (s DiskScenario) Run(ctx context.Context, opts Options) error {
	f, err := os.CreateTemp("", "kerno-chaos-disk-")
	if err != nil {
		return fmt.Errorf("create tmp file: %w", err)
	}
	path := f.Name()
	defer func() {
		_ = f.Close()
		_ = os.Remove(path)
	}()

	blockSize := blockSizeFromIntensity(opts.Intensity)
	block := make([]byte, blockSize)
	if _, err := rand.Read(block); err != nil {
		return fmt.Errorf("seed block: %w", err)
	}

	const truncateThreshold = 16 << 20 // 16 MB
	var written int64
	var ops uint64

	fmt.Fprintf(opts.Out, "    fsyncing %d-byte blocks into %s\n", blockSize, path)

	for ctx.Err() == nil {
		n, err := f.Write(block)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if err := f.Sync(); err != nil {
			return fmt.Errorf("sync: %w", err)
		}
		ops++
		written += int64(n)

		if written >= truncateThreshold {
			if err := f.Truncate(0); err != nil {
				return fmt.Errorf("truncate: %w", err)
			}
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return fmt.Errorf("seek: %w", err)
			}
			written = 0
		}
	}

	fmt.Fprintf(opts.Out, "    completed %d write+fsync operations\n", ops)
	return nil
}

func blockSizeFromIntensity(intensity Intensity) int {
	switch intensity {
	case IntensityLow:
		return 4096
	case IntensityHigh:
		return 64 * 1024
	default:
		return 16 * 1024
	}
}
