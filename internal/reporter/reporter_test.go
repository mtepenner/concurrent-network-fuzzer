package reporter

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"testing"
)

func newTempReporter(t *testing.T) (*JSONReporter, string) {
	t.Helper()
	f, err := os.CreateTemp("", "fuzzer-test-*.jsonl")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	name := f.Name()
	f.Close()
	// Remove so NewJSONReporter creates it fresh
	os.Remove(name)

	rep, err := NewJSONReporter(name)
	if err != nil {
		t.Fatalf("NewJSONReporter: %v", err)
	}
	return rep, name
}

func readResults(t *testing.T, path string) []FuzzResult {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open results: %v", err)
	}
	defer f.Close()

	var results []FuzzResult
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r FuzzResult
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		results = append(results, r)
	}
	return results
}

func TestJSONReporter_ClosedPortNotWritten(t *testing.T) {
	rep, path := newTempReporter(t)
	defer os.Remove(path)

	if err := rep.Report(FuzzResult{Port: 80, Open: false}); err != nil {
		t.Fatalf("Report: %v", err)
	}
	rep.Close()

	results := readResults(t, path)
	if len(results) != 0 {
		t.Errorf("expected 0 results for closed port, got %d", len(results))
	}
}

func TestJSONReporter_OpenPortWritten(t *testing.T) {
	rep, path := newTempReporter(t)
	defer os.Remove(path)

	want := FuzzResult{Port: 443, Open: true, Banner: "SSH-2.0", Crashed: false}
	if err := rep.Report(want); err != nil {
		t.Fatalf("Report: %v", err)
	}
	rep.Close()

	results := readResults(t, path)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	got := results[0]
	if got.Port != want.Port || got.Open != want.Open || got.Banner != want.Banner {
		t.Errorf("result mismatch: got %+v, want %+v", got, want)
	}
}

func TestJSONReporter_CrashedPortWritten(t *testing.T) {
	rep, path := newTempReporter(t)
	defer os.Remove(path)

	want := FuzzResult{Port: 21, Open: true, Crashed: true, ErrorMsg: "connection reset"}
	if err := rep.Report(want); err != nil {
		t.Fatalf("Report: %v", err)
	}
	rep.Close()

	results := readResults(t, path)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	got := results[0]
	if !got.Crashed || got.ErrorMsg != want.ErrorMsg {
		t.Errorf("result mismatch: got %+v, want %+v", got, want)
	}
}

func TestJSONReporter_ConcurrentWrites(t *testing.T) {
	rep, path := newTempReporter(t)
	defer os.Remove(path)

	const n = 100
	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			rep.Report(FuzzResult{Port: port, Open: true})
		}(i)
	}
	wg.Wait()
	rep.Close()

	results := readResults(t, path)
	if len(results) != n {
		t.Errorf("expected %d results, got %d", n, len(results))
	}
}

func TestJSONReporter_MultipleOpenPorts(t *testing.T) {
	rep, path := newTempReporter(t)
	defer os.Remove(path)

	for _, port := range []int{22, 80, 443} {
		rep.Report(FuzzResult{Port: port, Open: true})
	}
	rep.Close()

	results := readResults(t, path)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}
