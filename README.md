# Plagicheck

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: GPL v2](https://img.shields.io/badge/License-GPL%20v2-blue.svg)](https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/Software-Transparency-Foundation/stf-plagicheck)

A high-performance code plagiarism detection tool using winnowing fingerprints (WFP) and snippet matching techniques. Plagicheck scans source code files and directories against the osskb-core-open-dataset to identify potential code reuse.

## Features

- Full file matching detection
- Code snippet matching with line range identification
- Support for individual files, directories, and WFP files
- JSON output format for easy integration
- Debug mode for detailed processing information

## Installation

### Prerequisites

#### SCANOSS LDB
The LDB binary and shared library (libldb.so) must be installed on your system. For detailed installation instructions, please refer to:
https://github.com/scanoss/ldb/blob/master/README.md

#### SCANOSS Snippet Library
This project requires the `libsnippets_wrapper.a` static library for snippet scanning functionality. The library is included in the `deps/` directory but can also be built from source.

To build the library from source, follow the instructions at:
https://github.com/scanoss/engine/blob/main/go-wrapper/README.md

**Note:** The pre-compiled library in `deps/libsnippets_wrapper.a` is provided for convenience but may need to be rebuilt for your specific platform.

### Open Knowledge Base Dataset

To scan code, the **osskb-core-open-dataset** dataset must be available under the directory `/var/lib/ldb/`.

You can download the dataset from:
- FTP: ftp://osskb.st.foundation
- Web: http://osskb.st.foundation

**Note:** The total disk space required is approximately 1.2TB.

### Verifying LDB Installation

Once you have downloaded the knowledge base, verify the LDB directory structure:

```bash
ls -l /var/lib/ldb/
```

Expected output:
```
drwxr-xr-x  4 user user 4.0K Oct 27 21:01 osskbopen
```

Ensure the LDB binary is available in your PATH:

```bash
which ldb
```

Expected output:
```
/usr/bin/ldb
```

Test LDB functionality with the following command:

```bash
echo "select from osskbopen/file-url key 00fffff25afaa0d78ff1c6f41ba7f965 csv hex 16" | ldb
```

Expected output:
```
00fffff25afaa0d78ff1c6f41ba7f965,Source/AccelByteUe4Sdk/Private/Core/AccelByteServerCredentials.cpp,https://github.com/accelbyte/accelbyte-unreal-sdk-plugin/archive/24.3.0.zip,52
```

If you see this result, you are ready to proceed with building Plagicheck.
### Building Plagicheck from Source

```bash
make build
```

### Install to GOPATH

```bash
make install
```

## Usage

### Scan a File or Directory

Scan a single file (WFP is generated automatically):
```bash
plagicheck myfile.go
```

Scan an entire directory:
```bash
plagicheck ./src
```

Scan a pre-generated WFP file:
```bash
plagicheck myproject.wfp
```

Use multiple threads for faster processing:
```bash
plagicheck -T 8 ./src
```

Enable debug mode for detailed information:
```bash
plagicheck -d myfile.go
```

### Generate WFP Only

Generate WFP and output to stdout:
```bash
plagicheck -fp myfile.go
```

Generate WFP and save to file:
```bash
plagicheck -fp --output myproject.wfp ./src
```

### Configure Hit Threshold

Require at least 10 hits for valid snippet match:
```bash
plagicheck --min-hits 10 myfile.wfp
```

### Version Information

Display version and commit information:
```bash
plagicheck --version
```

## Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-fp` | Generate WFP from file or directory (output only, no scan) | - |
| `--output <file>` | Output file for generated WFP | stdout |
| `--min-hits <N>` | Minimum number of hits required for valid snippet match | 3 |
| `-T <threads>` | Number of parallel threads for processing files | 3 |
| `-d` | Enable debug mode (show detailed processing information) | false |
| `--version` | Show version information | - |

## Output Format

Plagicheck outputs results in JSON format, making it easy to integrate with other tools and workflows.

### Match Types

#### Full File Match
An exact file match found in the knowledge base:
```json
{
  "match_type": "full_file",
  "instances": 52,
  "reference_url": "https://github.com/accelbyte/accelbyte-unreal-sdk-plugin/archive/24.3.0.zip",
  "reference_file": "Source/AccelByteUe4Sdk/Private/Core/AccelByteServerCredentials.cpp"
}
```

#### Code Snippet Match
A partial code match with line ranges indicating where the matching code is located:
```json
{
  "match_type": "code_snippet",
  "target_lines": "52-80",
  "ref_file_lines": "42-70",
  "instances": 52,
  "reference_url": "https://github.com/example/repository",
  "reference_file": "path/to/file.cpp"
}
```

**Fields explanation:**
- `target_lines`: Line range in your scanned file where the match was found
- `ref_file_lines`: Line range in the reference file that matches your code
- `instances`: Number of times this file appears in the knowledge base

#### No Match
No match found in the knowledge base:
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

## How It Works

### Winnowing Fingerprints (WFP)

Plagicheck uses the winnowing algorithm to generate fingerprints of source code files. This technique:

1. **Normalizes the code** - Removes whitespace and comments to focus on code structure.
2. **Generates hashes** - Creates hash values for sliding windows of code.
3. **Selects fingerprints** - Uses the winnowing algorithm to select a minimal set of representative hashes.
Please refer to https://github.com/scanoss/wfp for more details.

4. **Compares against database** - Matches fingerprints against the STF Open Knowledge Base.

### Matching Logic

#### Full File Matching
When the MD5 hash of the entire file matches a file in the knowledge base, a full file match is reported.

#### Snippet Matching
For partial matches, Plagicheck:
- Identifies matching hash sequences between your code and the knowledge base
- Calculates line ranges where matches occur
- Merges nearby matches (separated by less than 3 lines)
- Filters out single-line matches
- Validates matches have at least the minimum number of hits (configurable with `--min-hits`)

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.
Contributions to the dataset are also accepted in this repository.

## License

**SPDX-License-Identifier:** GPL-2.0

**Copyright (C) 2025 Fundación Para La Transparencia del Software - STF**

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later version.
