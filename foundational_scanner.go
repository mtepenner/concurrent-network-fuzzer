package main

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// worker processes ports from the 'ports' channel and sends results to the 'results' channel.
func worker(ports, results chan int, wg *sync.WaitGroup, target string) {
	defer wg.Done()
	
	for p := range ports {
		address := fmt.Sprintf("%s:%d", target, p)
		
		// DialTimeout is crucial. Without it, your scanner will hang on dropped packets.
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err != nil {
			results <- 0 // 0 indicates the port is closed or filtered
			continue
		}
		
		conn.Close()
		results <- p // Send back the open port number
	}
}

func main() {
	// target := "scanme.nmap.org" // Good for safe internet testing
	target := "127.0.0.1"          // Localhost for rapid testing
	
	numWorkers := 100 
	portsToScan := 1024

	// Create buffered channels
	ports := make(chan int, numWorkers)
	results := make(chan int)
	var openPorts []int
	
	// WaitGroup ensures the main function waits for all workers to finish
	var wg sync.WaitGroup

	// 1. Spin up the worker pool
	for i := 0; i < cap(ports); i++ {
		wg.Add(1)
		go worker(ports, results, &wg, target)
	}

	// 2. Feed work to the workers in a separate goroutine
	// This prevents the main thread from blocking while writing to the channel
	go func() {
		for i := 1; i <= portsToScan; i++ {
			ports <- i
		}
		close(ports) // Close the channel to signal workers to exit when the queue is empty
	}()

	// 3. Close the results channel ONLY when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Collect results from the workers
	for port := range results {
		if port != 0 {
			openPorts = append(openPorts, port)
		}
	}

	// Sort and print the results
	sort.Ints(openPorts)
	fmt.Printf("Scan completed on %s:\n", target)
	for _, port := range openPorts {
		fmt.Printf("Port %d is open\n", port)
	}
}
