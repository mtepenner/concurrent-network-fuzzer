package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mtepenner/concerrent-network-fuzzer/internal/mutator"
	"github.com/mtepenner/concerrent-network-fuzzer/internal/reporter"
	"github.com/mtepenner/concerrent-network-fuzzer/internal/scanner"
)

func main() {
	// Parse CLI arguments
	target := flag.String("target", "127.0.0.1", "Target IP address to fuzz")
	ports := flag.Int("ports", 1024, "Number of ports to scan/fuzz (e.g., 1024 means ports 1-1024)")
	workers := flag.Int("workers", 50, "Number of concurrent goroutines")
	outFile := flag.String("out", "results.jsonl", "File path for JSONL output")
	flag.Parse()

	fmt.Printf("Starting Fuzzing Run against %s (Ports: 1-%d, Workers: %d)\n", *target, *ports, *workers)

	// Initialize the JSON file reporter
	rep, err := reporter.NewJSONReporter(*outFile)
	if err != nil {
		log.Fatalf("Failed to initialize reporter: %v", err)
	}
	defer rep.Close()

	// Initialize the Payload Mutator (2KB buffer overflow)
	mut := mutator.NewBufferOverflowMutator(2048)

	// Build and run the Fuzzer
	fuzzer := scanner.NewFuzzer(*target, *ports, *workers, mut, rep)
	fuzzer.Run()

	fmt.Printf("Fuzzing Complete. Results saved to %s\n", *outFile)
}
