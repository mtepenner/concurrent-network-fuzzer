# 🚀 Concurrent Network Fuzzer

A high-performance, concurrent network vulnerability scanner and fuzzer written in Go. Designed to test network services for stability issues, this tool leverages Go's concurrency model to quickly scan multiple ports, grab banners, and deliver custom payload mutations (such as buffer overflows and format string attacks) to identify potential crashes.

## 📑 Table of Contents
- [Features](#-features)
- [Technologies Used](#-technologies-used)
- [Installation](#-installation)
- [Usage](#-usage)
- [Architecture](#-architecture)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features
* **Concurrent Execution:** Utilizes worker pools to simultaneously scan and fuzz multiple ports, significantly reducing testing time.
* **Custom Payload Mutators:** Easily swap or add payload generators using the `Mutator` interface. Built-in mutators include:
  * `BufferOverflowMutator`: Generates variable-length repeating byte payloads.
  * `FormatStringMutator`: Generates repeated format string specifiers (e.g., `%x`, `%n`).
* **Health Monitoring:** Automatically monitors the target connection after payload delivery to detect crashes or timeouts.
* **Banner Grabbing:** Captures and logs service banners during the initial connection phase.
* **Structured Reporting:** Thread-safe JSONL reporting ensures clean, parsable logs of open ports and fuzzing results without data corruption.

## 🛠️ Technologies Used
* **Language:** Go 1.21
* **Standard Libraries:** `net`, `sync`, `flag`, `encoding/json`

## 📦 Installation

Ensure you have Go 1.21 or later installed on your machine. 

1. Clone the repository:
   ```bash
   git clone https://github.com/mtepenner/concurrent-network-fuzzer.git
   cd concurrent-network-fuzzer
   ```

2. Build the binary using the provided Makefile:
   ```bash
   make build
   ```
   *This compiles the executable and places it in the `bin/` directory.*

## 💻 Usage

You can run the fuzzer directly via the Makefile or by executing the compiled binary. 

**Using Makefile:**
```bash
make run
```
*By default, this targets `127.0.0.1` and scans up to port `1024`.*

**Using the Binary:**
```bash
./bin/fuzzer [options]
```

### Command-Line Flags
| Flag | Default | Description |
| :--- | :--- | :--- |
| `--target` | `127.0.0.1` | Target IP address to fuzz. |
| `--ports` | `1024` | Number of ports to scan/fuzz sequentially (e.g., `1024` means ports 1-1024). |
| `--workers`| `50` | Number of concurrent goroutines for scanning. |
| `--out` | `results.jsonl` | File path for the structured JSON Lines output. |

### Example
Scan a local server on the first 5000 ports using 100 concurrent workers, saving results to `audit.jsonl`:
```bash
./bin/fuzzer --target 192.168.1.10 --ports 5000 --workers 100 --out audit.jsonl
```

## 🏗️ Architecture
The project is structured into three primary internal packages:
1. **`scanner`:** Manages the worker pool, TCP connections, banner grabbing, and timeout configurations.
2. **`mutator`:** Defines the `Mutator` interface and implements specific payload generation logic (Buffer Overflows, Format Strings).
3. **`reporter`:** Handles thread-safe disk writes (`os.O_APPEND`), ensuring that asynchronous fuzzing results from the workers are correctly serialized into JSON lines.

## 🤝 Contributing
Contributions, issues, and feature requests are welcome! 
1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License
Distributed under the MIT License. See `LICENSE` for more information.
