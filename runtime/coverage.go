package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Position represents a source location in a template file
type Position struct {
	Line uint32 `json:"line"`
	Col  uint32 `json:"col"`
}

// CoveragePoint represents a single coverage measurement
type CoveragePoint struct {
	Line uint32 `json:"line"`
	Col  uint32 `json:"col"`
	Hits uint32 `json:"hits"`
	Type string `json:"type"`
}

// CoverageProfile represents the JSON coverage output format
type CoverageProfile struct {
	Version string                     `json:"version"`
	Mode    string                     `json:"mode"`
	Files   map[string][]CoveragePoint `json:"files"`
}

// CoverageRegistry tracks coverage data during test execution
type CoverageRegistry struct {
	mu       sync.Mutex
	files    map[string]map[Position]uint32 // filename → position → hit count
	coverDir string                         // captured at init time, used at flush time
}

var (
	coverageRegistry *CoverageRegistry
	coverageOnce     sync.Once
)

// EnableCoverage initializes the coverage registry for non-test use cases (e.g. servers).
// For test binaries, use RunWithCoverage instead.
// Unlike the old EnableCoverageForTesting(), this requires TEMPLCOVERDIR to be set —
// it won't initialize the registry without a target directory.
func EnableCoverage() {
	dir := os.Getenv("TEMPLCOVERDIR")
	if dir == "" {
		return
	}
	coverageOnce.Do(func() {
		coverageRegistry = &CoverageRegistry{
			files:    make(map[string]map[Position]uint32),
			coverDir: dir,
		}
	})
}

// TestRunner is implemented by *testing.M (which has a Run() method).
type TestRunner interface {
	Run() int
}

// RunWithCoverage wraps m.Run() with coverage lifecycle management.
// If TEMPLCOVERDIR is not set, it calls m.Run() directly with zero overhead.
// Safe to leave in permanently — it's a no-op without TEMPLCOVERDIR.
func RunWithCoverage(m TestRunner) int {
	dir := os.Getenv("TEMPLCOVERDIR")
	if dir == "" {
		return m.Run()
	}

	coverageOnce.Do(func() {
		coverageRegistry = &CoverageRegistry{
			files:    make(map[string]map[Position]uint32),
			coverDir: dir,
		}
	})

	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			_ = coverageRegistry.Flush()
			os.Exit(1)
		case <-done:
			return
		}
	}()

	code := m.Run()

	signal.Stop(sigChan)
	close(done)
	_ = coverageRegistry.Flush()
	return code
}

// CoverageSnapshot returns a deep copy of current coverage data.
// Returns nil if coverage is not enabled. Thread-safe.
func CoverageSnapshot() map[string][]CoveragePoint {
	if coverageRegistry == nil {
		return nil
	}
	coverageRegistry.mu.Lock()
	defer coverageRegistry.mu.Unlock()

	result := make(map[string][]CoveragePoint, len(coverageRegistry.files))
	for filename, positions := range coverageRegistry.files {
		points := make([]CoveragePoint, 0, len(positions))
		for pos, hits := range positions {
			points = append(points, CoveragePoint{
				Line: pos.Line,
				Col:  pos.Col,
				Hits: hits,
			})
		}
		result[filename] = points
	}
	return result
}

// Record increments the hit count for a coverage point
func (r *CoverageRegistry) Record(filename string, line, col uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.files[filename] == nil {
		r.files[filename] = make(map[Position]uint32)
	}

	pos := Position{Line: line, Col: col}
	r.files[filename][pos]++
}

// CoverageTrack records that a coverage point was executed
// Called by generated template code when coverage is enabled
func CoverageTrack(filename string, line, col uint32) {
	if coverageRegistry == nil {
		return // No-op if coverage disabled
	}
	coverageRegistry.Record(filename, line, col)
}

// WriteProfile writes coverage data to a JSON file
func (r *CoverageRegistry) WriteProfile(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	profile := CoverageProfile{
		Version: "1",
		Mode:    "count",
		Files:   make(map[string][]CoveragePoint),
	}

	// Convert internal map to slice format
	for filename, positions := range r.files {
		points := make([]CoveragePoint, 0, len(positions))
		for pos, hits := range positions {
			points = append(points, CoveragePoint{
				Line: pos.Line,
				Col:  pos.Col,
				Hits: hits,
				Type: "expression", // Default type for now
			})
		}
		profile.Files[filename] = points
	}

	// Write JSON
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	return nil
}

// Flush writes the coverage profile to disk
func (r *CoverageRegistry) Flush() error {
	if r.coverDir == "" {
		return nil
	}

	if err := os.MkdirAll(r.coverDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("templ-%d-%d.json", os.Getpid(), time.Now().UnixNano())
	path := filepath.Join(r.coverDir, filename)

	return r.WriteProfile(path)
}

// FlushCoverage explicitly flushes coverage data to disk
// Tests should call this in cleanup to ensure profiles are written
func FlushCoverage() error {
	if coverageRegistry == nil {
		return nil
	}
	return coverageRegistry.Flush()
}
