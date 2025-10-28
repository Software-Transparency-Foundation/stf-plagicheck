# libsnippets_wrapper

`libsnippets_wrapper` is a static library that provides an interface to the snippet scanning capabilities from the SCANOSS engine. This library acts as a bridge between the Go code in this project and the underlying C implementation of the SCANOSS scanning engine.

## Purpose

The library wraps the core snippet scanning functionality, allowing the OSSKB Engine to:

- Scan source code files for snippet matches against the knowledge base
- Perform fingerprint-based comparisons using Winnowing algorithm hashes
- Identify file matches, snippet matches, and binary file matches
- Retrieve detailed match information including line ranges and hit counts

## Building the Library

For detailed instructions on how to build `libsnippets_wrapper.a`, please refer to the official SCANOSS documentation:

**[SCANOSS Engine Go Wrapper Build Guide](https://github.com/scanoss/engine/blob/main/go-wrapper/README.md)**

### Files in this Directory

- **`snippets_wrapper.h`** - C header file defining the wrapper API interface
- **`wfp_scanner.go`** - Go implementation that interfaces with the C library via CGO
- **`cmd/main.go`** - Command-line tool for testing WFP scanning functionality
- **`libsnippets_wrapper.a`** - Static library binary (must be built separately)

## API Overview

The wrapper provides the following main functions:

- `snippets_wrapper_init()` - Initialize the scanning engine with a knowledge base
- `snippets_wrapper_scan()` - Perform a scan on WFP (Winnowing Fingerprint) data
- `snippets_wrapper_free_result()` - Free memory allocated for scan results
- `snippets_wrapper_cleanup()` - Clean up resources and close the scanning engine

## License

The SCANOSS Open Source Engine is released under the GPL 2.0 license. Please check the LICENSE file for more information.

**Copyright (C) 2018-2025 SCANOSS.COM**