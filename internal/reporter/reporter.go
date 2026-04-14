package reporter

// FuzzResult holds the structured outcome of a single port fuzzing attempt.
type FuzzResult struct {
	Port     int    `json:"port"`
	Open     bool   `json:"open"`
	Banner   string `json:"banner,omitempty"`
	Crashed  bool   `json:"crashed"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

// Reporter handles outputting the results to various formats (JSON, stdout, etc.)
type Reporter interface {
	Report(result FuzzResult) error
	Close() error
}
