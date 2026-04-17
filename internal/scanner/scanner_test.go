package scanner

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/mtepenner/concurrent-network-fuzzer/internal/reporter"
)

// --- Helpers ---

// fixedMutator always returns the same payload.
type fixedMutator struct {
	payload []byte
}

func (m *fixedMutator) Generate() []byte { return m.payload }
func (m *fixedMutator) Name() string     { return "fixed" }

// memReporter collects results in memory, thread-safely.
type memReporter struct {
	mu      sync.Mutex
	results []reporter.FuzzResult
}

func (r *memReporter) Report(result reporter.FuzzResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, result)
	return nil
}

func (r *memReporter) Close() error { return nil }

func (r *memReporter) all() []reporter.FuzzResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]reporter.FuzzResult, len(r.results))
	copy(cp, r.results)
	return cp
}

// startEchoServer starts a TCP server that echoes all data back.
// Returns the port it's listening on and a function to stop it.
func startEchoServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go io.Copy(conn, conn)
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

// startBannerServer starts a TCP server that sends a banner then reads/discards data.
func startBannerServer(t *testing.T, banner string) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				c.Write([]byte(banner))
				io.Copy(io.Discard, c)
			}(conn)
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

// startCloseOnConnectServer accepts a connection and immediately closes it.
func startCloseOnConnectServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

// --- Tests ---

func TestFuzzer_ClosedPortNotOpen(t *testing.T) {
	// Bind a port, record its number, then close it so the OS stops listening.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	// Small sleep to let the OS reclaim the port before we try to connect.
	time.Sleep(10 * time.Millisecond)

	// Call fuzzPort directly so we only probe the one closed port.
	rep := &memReporter{}
	mut := &fixedMutator{payload: []byte("AAAA")}
	f := NewFuzzer("127.0.0.1", port, 1, mut, rep)
	result := f.fuzzPort(port)

	if result.Open {
		t.Errorf("expected port %d to be closed, but got Open=true", port)
	}
	if result.Port != port {
		t.Errorf("expected result for port %d, got %d", port, result.Port)
	}
}

func TestFuzzer_OpenPortDetected(t *testing.T) {
	port, stop := startEchoServer(t)
	defer stop()

	rep := &memReporter{}
	mut := &fixedMutator{payload: []byte("hello")}
	f := NewFuzzer("127.0.0.1", 1, 1, mut, rep)
	result := f.fuzzPort(port)

	if !result.Open {
		t.Errorf("expected port %d to be open", port)
	}
	if result.Port != port {
		t.Errorf("expected result for port %d, got %d", port, result.Port)
	}
}

func TestFuzzer_BannerCaptured(t *testing.T) {
	const banner = "SSH-2.0-OpenSSH_8.9"
	port, stop := startBannerServer(t, banner)
	defer stop()

	rep := &memReporter{}
	mut := &fixedMutator{payload: []byte("payload")}
	f := NewFuzzer("127.0.0.1", 1, 1, mut, rep)
	result := f.fuzzPort(port)

	if result.Banner != banner {
		t.Errorf("expected banner %q, got %q", banner, result.Banner)
	}
}

func TestFuzzer_CrashDetectedOnConnectionReset(t *testing.T) {
	port, stop := startCloseOnConnectServer(t)
	defer stop()

	rep := &memReporter{}
	mut := &fixedMutator{payload: []byte("AAAAAAAAAA")}
	f := NewFuzzer("127.0.0.1", 1, 1, mut, rep)
	result := f.fuzzPort(port)

	// Server closes connection immediately after connect, so writing the
	// payload or reading the response will produce an error that should
	// be reported as a crash or write failure.
	if !result.Crashed && result.ErrorMsg == "" {
		t.Error("expected Crashed=true or a non-empty ErrorMsg for a forcibly closed connection")
	}
}

func TestFuzzer_ConcurrentWorkers(t *testing.T) {
	// Start several echo servers and scan all their ports concurrently.
	const numServers = 5
	ports := make([]int, numServers)
	stops := make([]func(), numServers)
	for i := range ports {
		p, s := startEchoServer(t)
		ports[i] = p
		stops[i] = s
	}
	defer func() {
		for _, s := range stops {
			s()
		}
	}()

	maxPort := 0
	for _, p := range ports {
		if p > maxPort {
			maxPort = p
		}
	}

	rep := &memReporter{}
	mut := &fixedMutator{payload: []byte("fuzz")}
	f := &Fuzzer{Target: "127.0.0.1", Ports: maxPort, Workers: 20, Mutator: mut, Reporter: rep}
	f.Run()

	// All results for our servers' ports must be present and open.
	byPort := map[int]reporter.FuzzResult{}
	for _, r := range rep.all() {
		byPort[r.Port] = r
	}
	for _, p := range ports {
		r, ok := byPort[p]
		if !ok {
			t.Errorf("no result for port %d", p)
			continue
		}
		if !r.Open {
			t.Errorf("port %d should be open", p)
		}
	}
}

func TestFuzzer_Run_AllPortsScanned(t *testing.T) {
	// Run against a range of ports on localhost and verify every port in
	// [1, Ports] produces exactly one result.
	const portCount = 10
	rep := &memReporter{}
	mut := &fixedMutator{payload: []byte("A")}
	f := &Fuzzer{Target: "127.0.0.1", Ports: portCount, Workers: 5, Mutator: mut, Reporter: rep}
	f.Run()

	results := rep.all()
	if len(results) != portCount {
		t.Errorf("expected %d results (one per port), got %d", portCount, len(results))
	}

	seen := map[int]bool{}
	for _, r := range results {
		if r.Port < 1 || r.Port > portCount {
			t.Errorf("result has out-of-range port %d", r.Port)
		}
		if seen[r.Port] {
			t.Errorf("duplicate result for port %d", r.Port)
		}
		seen[r.Port] = true
	}
	for p := 1; p <= portCount; p++ {
		if !seen[p] {
			t.Errorf("missing result for port %d", p)
		}
	}
}

func TestNewFuzzer_Fields(t *testing.T) {
	rep := &memReporter{}
	mut := &fixedMutator{}
	f := NewFuzzer("10.0.0.1", 100, 10, mut, rep)
	if f.Target != "10.0.0.1" {
		t.Errorf("Target: got %q", f.Target)
	}
	if f.Ports != 100 {
		t.Errorf("Ports: got %d", f.Ports)
	}
	if f.Workers != 10 {
		t.Errorf("Workers: got %d", f.Workers)
	}
	_ = fmt.Sprintf("%v", f) // just ensure it doesn't panic
}
