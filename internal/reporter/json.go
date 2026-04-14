package reporter

import (
	"encoding/json"
	"os"
	"sync"
)

type JSONReporter struct {
	file *os.File
	mu   sync.Mutex // Mutex prevents concurrent workers from corrupting the file write
}

func NewJSONReporter(filename string) (*JSONReporter, error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONReporter{file: f}, nil
}

func (j *JSONReporter) Report(result FuzzResult) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// For cleaner logs, we only report ports that are actually open
	if !result.Open {
		return nil
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = j.file.Write(data)
	return err
}

func (j *JSONReporter) Close() error {
	return j.file.Close()
}
