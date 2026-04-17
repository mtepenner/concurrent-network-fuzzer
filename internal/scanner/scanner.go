package scanner

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/mtepenner/concurrent-network-fuzzer/internal/mutator"
	"github.com/mtepenner/concurrent-network-fuzzer/internal/reporter"
)

type Fuzzer struct {
	Target   string
	Ports    int
	Workers  int
	Mutator  mutator.Mutator
	Reporter reporter.Reporter
}

func NewFuzzer(target string, ports, workers int, mut mutator.Mutator, rep reporter.Reporter) *Fuzzer {
	return &Fuzzer{
		Target:   target,
		Ports:    ports,
		Workers:  workers,
		Mutator:  mut,
		Reporter: rep,
	}
}

func (f *Fuzzer) Run() {
	portsChan := make(chan int, f.Workers)
	var wg sync.WaitGroup

	// Spin up worker pool
	for i := 0; i < f.Workers; i++ {
		wg.Add(1)
		go f.worker(portsChan, &wg)
	}

	// Feed the work queue
	for i := 1; i <= f.Ports; i++ {
		portsChan <- i
	}
	close(portsChan)

	// Wait for all workers to finish processing
	wg.Wait()
}

func (f *Fuzzer) worker(ports <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range ports {
		result := f.fuzzPort(p)
		if err := f.Reporter.Report(result); err != nil {
			log.Printf("reporter error on port %d: %v", p, err)
		}
	}
}

func (f *Fuzzer) fuzzPort(port int) reporter.FuzzResult {
	result := reporter.FuzzResult{Port: port}
	address := fmt.Sprintf("%s:%d", f.Target, port)

	// 1. Initial Connection
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		result.Open = false
		return result
	}
	defer conn.Close()
	result.Open = true

	// 2. Read Banner
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	if n > 0 {
		result.Banner = string(buffer[:n])
	}

	// 3. Deliver Payload
	payload := f.Mutator.Generate()
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Write(payload)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("Failed to write payload: %v", err)
		return result
	}

	// 4. Monitor Health
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buffer)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// Expected behavior, the server ignored the payload
		} else {
			result.Crashed = true
			result.ErrorMsg = err.Error()
		}
	}

	return result
}
