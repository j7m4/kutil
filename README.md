# kutil

A lightweight Kubernetes utility toolkit.

Work in progress: limited to working with kubernetes contexts.

## Table of Contents

- [User Guide](#user-guide)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Commands](#commands)
  - [Examples](#examples)
- [Developer Guide](#developer-guide)
  - [Project Structure](#project-structure)
  - [Building from Source](#building-from-source)
  - [Development Workflow](#development-workflow)
  - [Adding New Commands](#adding-new-commands)
  - [Testing](#testing)

---

## User Guide

`kctx` is a command-line tool that simplifies Kubernetes context management with powerful regex-based operations.

### Installation

#### From Source

```bash
# Clone the repository
git clone <repository-url>
cd kutil

# Build the binary
make build

# Install to system (requires sudo)
make install
```

The binary will be installed to `/usr/local/bin/kctx`.

#### Manual Installation

```bash
go build -o kctx ./cmd/kctx
sudo mv kctx /usr/local/bin/
sudo chmod +x /usr/local/bin/kctx
```

### Usage

```bash
kctx <command> [flags] [arguments]
```

### Commands

#### `ls` - List Contexts

Lists all Kubernetes contexts in your kubeconfig. The current context is marked with an asterisk (*).

```bash
kctx ls
```

**Output:**
```
* production-cluster
  staging-cluster
  dev-cluster
```

#### `grep` - Filter Contexts

Filter and display contexts matching a regex pattern.

```bash
kctx grep REGEX [flags]
```

**Flags:**
- `-v, --invert-match` - Show contexts that do NOT match the pattern

**Examples:**
```bash
# Find all production contexts
kctx grep "prod"

# Find contexts containing "dev" or "test"
kctx grep "dev|test"

# Find contexts NOT containing "staging"
kctx grep -v "staging"
```

#### `tr` - Transform Context Names

Transform context names using regex patterns. This command modifies your kubeconfig file.

```bash
# Replace mode
kctx tr INPUT_REGEX REPLACEMENT_VALUE [flags]

# Delete mode
kctx tr -d DELETION_REGEX [flags]
```

**Flags:**
- `-d, --delete` - Delete matched regex from context names
- `-f, --force` - Apply changes without confirmation prompt

**Examples:**
```bash
# Replace "prod" with "production" in all context names
kctx tr "prod" "production"

# Remove "-eks" suffix from all contexts
kctx tr -d "-eks$"

# Replace cluster prefix without confirmation
kctx tr "^old-" "new-" -f

# Remove all numbers from context names
kctx tr -d "[0-9]+"
```

**Note:** The `tr` command will show you the proposed changes and ask for confirmation unless you use the `-f` flag.

#### `backup` - Backup Kubeconfig

Create a timestamped backup of your kubeconfig file.

```bash
kctx backup
```

**Output:**
```
Backup created: /Users/username/.kube/config_backup_202501081430
```

### Examples

#### Renaming Multiple Contexts

```bash
# List contexts to see what needs changing
kctx ls

# Preview changes
kctx tr "dev-cluster" "development-cluster"

# Output shows:
# dev-cluster-us-east-1 -> development-cluster-us-east-1
# dev-cluster-us-west-2 -> development-cluster-us-west-2
# Apply changes to 2 context(s)? [y/N]:
```

#### Safe Bulk Operations

```bash
# Always backup before bulk operations
kctx backup

# Make changes with confirmation
kctx tr "old-pattern" "new-pattern"
```

#### Finding Specific Contexts

```bash
# Find all AWS EKS contexts
kctx grep "eks"

# Find all contexts except production
kctx grep -v "prod"

# Filter by region
kctx grep "us-west"
```

---

## Developer Guide

### Building from Source

#### Prerequisites

- Go 1.25.1 or later
- Make (optional, but recommended)

#### Build Commands

```bash
# Download dependencies
make deps

# Build the kctx binary
make build

# Build and run tests
make test

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run all checks (format, vet, test)
make check

# Clean build artifacts
make clean
```

### Development Workflow

1. **Setup Development Environment**

```bash
# Install development tools
make setup
```

2. **Make Changes**

Edit code in `cmd/kctx/` or `internal/kctx/` directories.

3. **Test Changes**

```bash
# Format code
make fmt

# Run tests
make test

# Build locally
make build

# Test the binary
./bin/kctx ls
```

4. **Submit Changes**

```bash
# Run all checks before committing
make check
```

### Adding New Commands

To add a new command to `kctx`:

1. **Create Implementation** in `internal/kctx/`

```go
// internal/kctx/newcommand.go
package kctx

import (
    "fmt"
    "k8s.io/client-go/tools/clientcmd"
)

func NewCommandFunc() error {
    // Load kubeconfig
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    configOverrides := &clientcmd.ConfigOverrides{}
    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loadingRules,
        configOverrides,
    )

    rawConfig, err := clientConfig.RawConfig()
    if err != nil {
        return fmt.Errorf("error loading kubeconfig: %v", err)
    }

    // Your logic here

    return nil
}
```

2. **Register Command** in `cmd/kctx/main.go`

```go
// Add to main() function
newCmd := &cobra.Command{
    Use:   "newcommand",
    Short: "Brief description",
    Long:  "Detailed description",
    Args:  cobra.NoArgs,
    Run:   newCommandHandler,
}

// Add flags if needed
newCmd.Flags().StringP("option", "o", "", "Option description")

// Register with root command
rootCmd.AddCommand(newCmd)

// Add handler function
func newCommandHandler(cmd *cobra.Command, args []string) {
    if err := kctx.NewCommandFunc(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

3. **Test Your Command**

```bash
make build
./bin/kctx newcommand
```

### Testing

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

#### Writing Tests

Create test files alongside your implementation:

```go
// internal/kctx/newcommand_test.go
package kctx

import (
    "testing"
)

func TestNewCommandFunc(t *testing.T) {
    // Your test logic
}
```

### Code Style

This project follows standard Go conventions:

- Use `gofmt` for formatting (run `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Write clear, descriptive commit messages
- Add comments for exported functions and complex logic

### Dependencies

Key dependencies:

- **k8s.io/client-go** - Kubernetes Go client library for kubeconfig operations
- **github.com/spf13/cobra** - CLI framework for command structure
- **github.com/spf13/pflag** - POSIX/GNU-style flags

### Troubleshooting Development Issues

#### Module Issues

```bash
# Clean and re-download dependencies
go clean -modcache
make deps
```

#### Build Issues

```bash
# Clean and rebuild
make clean
make build
```

#### Permission Issues

```bash
# Make sure the binary is executable
chmod +x bin/kctx
```

### Contributing

When contributing:

1. Create a backup of your kubeconfig before testing: `kctx backup`
2. Test changes with various kubeconfig scenarios
3. Run `make check` before submitting
4. Update documentation for new features
5. Consider edge cases (empty contexts, invalid regex, etc.)
