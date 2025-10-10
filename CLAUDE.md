# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Local Beach is a Docker-based development environment for Neos CMS and Flow Framework. It's a CLI tool written in Go that manages local development instances using Docker containers (Nginx, PHP, Redis, and MySQL).

The tool provides commands to initialize, start, stop, and manage Neos/Flow projects locally, with features like resource synchronization with Beach cloud storage, HTTPS setup, and Docker container orchestration.

## Build and Development Commands

### Building the Project

**Full build (with asset generation):**
```bash
make
```
This runs:
- `rm -f assets/compiled.go` - removes compiled assets
- `go generate -v` - generates embedded assets using vfsgen
- `go install -v` - installs dependencies
- `go build -v -ldflags "-X github.com/flownative/localbeach/pkg/version.Version=dev" -o beach` - builds binary

**Quick compile (without regenerating assets):**
```bash
make compile
```
Use this during development when assets haven't changed.

### Testing

Currently no test files exist in the project. When running `go test ./...`, note that there's a known build failure in `cmd/beach/cmd/resource-upload.go:84:3` due to a Printf formatting directive in a logrus.Fatal call.

## Architecture

### Project Structure

- **`cmd/beach/cmd/`** - Cobra-based CLI commands (each command in its own file)
- **`pkg/beachsandbox/`** - Core sandbox abstraction that represents a Local Beach project instance
- **`pkg/exec/`** - Docker command execution wrapper
- **`pkg/path/`** - Platform-specific path handling (Darwin/Linux)
- **`pkg/version/`** - Version information
- **`assets/`** - Embedded template files (Docker Compose configs, env templates, etc.)

### Key Concepts

**BeachSandbox**: The central abstraction representing a Local Beach project. Key responsibilities:
- Project detection via `.localbeach.docker-compose.yaml` marker file
- Environment variable loading from `.localbeach.dist.env`, `.localbeach.env`, `.env` (in that order)
- Flow/Neos installation detection
- Docker Compose file path management

**Project Detection**: Commands traverse up from the current directory looking for `.localbeach.docker-compose.yaml` to find the project root (see `pkg/beachsandbox/helpers.go:detectProjectRootPath`).

**Asset Embedding**: Template files in `assets/` are compiled into the binary using vfsgen. During build, `go generate` runs `assets_generate.go` which embeds all files from `assets/` into `assets/compiled.go`.

### Docker Architecture

Local Beach uses a two-tier Docker setup:

1. **Global infrastructure** (started by `beach start`):
   - `local_beach_nginx` - Reverse proxy for all projects
   - `local_beach_database` - Shared MySQL database server
   - Managed by `assets/local-beach/docker-compose.yml`

2. **Project-specific containers**:
   - Defined in `.localbeach.docker-compose.yaml` per project
   - Each project gets its own PHP-FPM container and services
   - Projects access the database at `http://{project-name}.localbeach.net`

### Important File Locations

**macOS:**
- Base path: `~/Library/Application Support/Flownative/Local Beach/`

**Linux/Other:**
- Base path: `~/.Flownative/Local Beach/`

These paths are defined in `pkg/path/path_darwin.go` and `pkg/path/path_linux.go`.

## Common Development Patterns

### Adding a New Command

1. Create a new file in `cmd/beach/cmd/` (e.g., `my-command.go`)
2. Define a `cobra.Command` struct with Use, Short, Long, Args, and Run fields
3. Add `rootCmd.AddCommand(myCmd)` in the `init()` function
4. Implement the `handleMyCommandRun` function
5. Use `beachsandbox.GetActiveSandbox()` to get the current project context if needed
6. Use `pkg/exec.RunCommand()` or `pkg/exec.RunInteractiveCommand()` for Docker operations

### Environment Variable Handling

Environment files are loaded in order (later files override earlier ones):
1. `.localbeach.dist.env` - Committed defaults
2. `.localbeach.env` - Local overrides
3. `.env` - Additional local config

Variables are parsed and set via `loadLocalBeachEnvironment()` in `pkg/beachsandbox/helpers.go`.

### Resource Path Calculation

Flow/Neos persistent resources use a hash-based directory structure. The `getRelativePersistentResourcePathByHash()` function in `cmd/beach/cmd/helpers.go` converts resource hashes to their filesystem path structure.

### Docker Command Execution and TTY Detection

The `pkg/exec` package provides two methods for executing Docker commands:

- **`RunInteractiveCommand()`**: Connects stdin/stdout/stderr for interactive sessions
- **`RunCommand()`**: Captures output without connecting stdin

**TTY Detection Pattern** (see `cmd/beach/cmd/exec.go`):

Commands that need to work both interactively (user's terminal) and programmatically (automation tools, Claude Code) should detect TTY availability:

```go
// Use syscall.TIOCGETA ioctl for reliable TTY detection
var termios syscall.Termios
_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(),
    syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
isTTY := errno == 0
```

This is more reliable than `os.Stat().Mode() & os.ModeCharDevice` which can incorrectly report TTY availability.

**Docker Exec Flags**:
- Interactive (TTY available): Use `-t -i` flags (allocate pseudo-TTY + keep stdin open)
- Non-interactive: Omit all flags (Docker's `-t` flag requires a real TTY)

**Error Handling**: When using `RunCommand()` in non-TTY mode, print output before checking errors so users see command output even on failure.

## Version Management

The version is injected at build time via ldflags:
```
-ldflags "-X github.com/flownative/localbeach/pkg/version.Version=dev"
```

For releases, replace "dev" with the actual version number.
