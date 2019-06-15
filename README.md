# protod

[![](https://godoc.org/github.com/dennwc/protod?status.svg)](http://godoc.org/github.com/dennwc/protod)

Sometimes it is necessary to recover `.proto` files from binaries or memory dumps.

`protod` does exactly this - it finds protobuf descriptors in any binary files and writes them back as `.proto` text files.

**Supports:**
- `proto2` and `proto3`
- `message`s, `enum`s, `service`s
- Extraction from uncompressed file descriptors (used in C/C++, maybe others)

**Not supported yet:**
- Field or file options
- Extensions
- Compressed file descriptors (used in Go)
- Recovery without file descriptors

## Installation

Go 1.12+ is required.

```
go get -u github.com/dennwc/protod
go install github.com/dennwc/protod/cmd/protod 
```

## Usage

```
protod --out=./out some_binary
```

The tool will emit recovered `.proto` files to `./out` directory.

## License

**MIT** (based on [protod](https://github.com/sysdream/Protod) Python script by Sysdream)