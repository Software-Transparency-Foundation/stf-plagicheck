# Plagicheck

A high-performance code plagiarism detection tool using winnowing fingerprints (WFP) and snippet matching.

## Features

- **WFP Generation**: Generate winnowing fingerprints from files or directories
- **File Matching**: Detect exact file matches using MD5 hashing
- **Snippet Matching**: Identify code snippets with configurable hit thresholds
- **Flexible Scanning**: Scan .wfp files, individual files, or entire directories
- **Smart Filtering**: Filter out single-line matches and apply minimum hit requirements

## Installation

### From Source

```bash
make build
```

### Install to GOPATH

```bash
make install
```

## Usage

### Scan a file or directory

```bash
# Scan a file directly (generates WFP automatically)
plagicheck myfile.go

# Scan a directory
plagicheck ./src

# Scan a WFP file
plagicheck myproject.wfp
```

### Generate WFP only

```bash
# Generate WFP and output to stdout
plagicheck -fp myfile.go

# Generate WFP and save to file
plagicheck -fp --output myproject.wfp ./src
```

### Configure hit threshold

```bash
# Require at least 10 hits for valid snippet match
plagicheck --min-hits 10 myfile.wfp
```

### Version information

```bash
plagicheck --version
```

## Command Line Options

- `-fp`: Generate WFP from file or directory (output only, no scan)
- `--output <file>`: Output file for generated WFP (default: stdout)
- `--min-hits <N>`: Minimum number of hits required for valid snippet match (default: 3)
- `--version`: Show version information

## Match Types

### Full File Match
Exact file match found in the knowledge base:
```json
{
  "match_type": "full_file",
  "instances": 52,
  "reference_url": "https://github.com/...",
  "reference_file": "path/to/file.cpp"
}
```

### Code Snippet Match
Partial code match with line ranges:
```json
{
  "match_type": "code_snippet",
  "target_lines": "52-80",
  "source_lines": "42-70",
  "instances": 52,
  "reference_url": "https://github.com/...",
  "reference_file": "path/to/file.cpp",
  "hits": 15
}
```

### No Match
No match found:
```json
{
  "match_type": "no_match",
  "instances": 0,
  "reference_url": "",
  "reference_file": ""
}
```

## Development

### Running Tests

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make run-tests
```

### Code Quality

```bash
# Format code
make fmt

# Run go vet
make vet

# Run all checks (fmt + vet + test)
make check
```

### Cleaning

```bash
make clean
```

## Project Structure

```
.
├── cmd/           # Main application entry point
├── pkg/           # Core packages
│   ├── scan.go       # Scanning and matching logic
│   ├── winnowing.go  # WFP generation
│   └── *_test.go     # Unit tests
├── models/        # Data structures
├── deps/          # Dependencies (C wrapper)
├── test/          # Test files
├── Makefile       # Build automation
└── README.md      # This file
```

## Validation Rules

### Snippet Matches
- Must have at least `--min-hits` number of hits (default: 3)
- Ranges must span more than one line (single-line matches are filtered)
- Ranges are merged if separated by less than 3 lines

## License

Copyright (c) 2024
