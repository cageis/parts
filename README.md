# Parts

[![CI/CD Pipeline](https://github.com/cageis/parts/actions/workflows/ci.yml/badge.svg)](https://github.com/cageis/parts/actions/workflows/ci.yml)
[![Code Quality](https://github.com/cageis/parts/actions/workflows/quality.yml/badge.svg)](https://github.com/cageis/parts/actions/workflows/quality.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cageis/parts)](https://goreportcard.com/report/github.com/cageis/parts)
[![codecov](https://codecov.io/gh/cageis/parts/branch/main/graph/badge.svg)](https://codecov.io/gh/cageis/parts)

A Go utility for merging partial configuration files into an aggregate file. Designed primarily for managing SSH configurations by combining multiple partial config files into your main SSH config.

## Features

- **Idempotent Operations**: Run multiple times without causing file bloat or whitespace growth
- **Automatic Section Management**: Uses comment markers to manage the merged section
- **Original Content Preservation**: Your existing configuration remains untouched
- **Dynamic Updates**: Add/remove partial files and re-run to update the aggregate
- **Dry-Run Mode**: Preview changes without modifying files using `--dry-run` or `-n`
- **Remove Mode**: Clean removal of partials sections using `--remove` or `-r`
- **Multiple Comment Styles**: Support for various file types with appropriate comment syntax

## Installation

### From Source

```bash
git clone <repository-url>
cd parts
make build
```

The binary will be created at `bin/parts`.

### Install to GOPATH

```bash
make install
```

## Usage

### Basic Usage

```bash
# Build mode (default): Merge partials into aggregate file
parts [--dry-run|-n] <aggregate-file> <partials-directory> <comment-style>

# Remove mode: Remove partials section from aggregate file
parts --remove [--dry-run|-n] <aggregate-file> <comment-style>
```

### SSH Configuration Examples

```bash
# Build: Merge SSH config partials
parts ~/.ssh/config ~/.ssh/config.d "#"

# Remove: Remove partials section from SSH config
parts --remove ~/.ssh/config "#"

# Preview changes without modifying files (dry-run)
parts --dry-run ~/.ssh/config ~/.ssh/config.d "#"
parts --remove --dry-run ~/.ssh/config "#"
```

**Build mode** will:
1. Read all files from `~/.ssh/config.d/`
2. Merge them into your `~/.ssh/config` file
3. Place the merged content between enhanced comment markers:

**Remove mode** will:
1. Locate the partials section in your config file
2. Remove the entire section cleanly
3. Preserve all your original content

**Enhanced comment markers:**
   ```
   # ============================
   # PARTIALS>>>>>
   # ============================
   <content from partial files>
   # ============================
   # PARTIALS<<<<<
   # ============================
   ```

### Quick Start

For SSH configuration management:

```bash
make quickstart
```

This builds the binary and runs the SSH config merge using default paths.

### Examples

See practical usage examples for different file types:

```bash
make examples
```

This runs demonstrations of parts with various comment styles including SSH config, JavaScript, CSS, SQL, and Python files. Examples are located in the `./examples/` directory.

## How It Works

1. **Reads** the aggregate file and removes any existing managed section (between PARTIALS>>>>> and PARTIALS<<<<< markers)
2. **Scans** the partials directory for all files
3. **Appends** the content of each partial file between the comment markers  
4. **Writes** the updated aggregate file

The process is idempotent - running it multiple times produces identical results without accumulating whitespace or duplicate content.

## Example

### Directory Structure
```
~/.ssh/
├── config                    # Main SSH config
└── config.d/                # Partials directory
    ├── work-servers         # Work-related SSH hosts
    ├── personal-servers     # Personal SSH hosts
    └── development          # Development environment hosts
```

### Partial File Example (`~/.ssh/config.d/work-servers`)
```
Host work-db
    HostName db.company.com
    User dbadmin
    Port 2222

Host work-app
    HostName app.company.com  
    User deploy
```

### Result in `~/.ssh/config`
```
# Your existing SSH config...
Host personal
    HostName myserver.com
    User me

# ============================
# PARTIALS>>>>>
# ============================
Host work-db
    HostName db.company.com
    User dbadmin
    Port 2222

Host work-app
    HostName app.company.com  
    User deploy

Host dev-server
    HostName dev.company.com
    User developer
# ============================
# PARTIALS<<<<<
# ============================
```

### Comment Block Examples

Different file types get appropriate comment structures:

**CSS Files:**
```css
/*
/* PARTIALS>>>>>
*/
.btn { padding: 10px 20px; }
/*
/* PARTIALS<<<<<
*/
```

**SQL Files:**
```sql
-- ============================
-- PARTIALS>>>>>
-- ============================
CREATE TABLE products (id INT PRIMARY KEY);
-- ============================
-- PARTIALS<<<<<
-- ============================
```

## Development

### Building

```bash
make build        # Build the binary
make help         # Show all available commands
```

### Testing

```bash
make test         # Run all tests
make check        # Run format, vet, and tests
```

### Code Quality

```bash
make fmt          # Format code
make vet          # Run go vet
make clean        # Clean build artifacts
```

## Requirements

- Go 1.16 or later

## License

[Add your license here]

## Development & Contributing

See our [ROADMAP.md](./ROADMAP.md) for planned features and improvements.

### Contributing
- Check the roadmap for high-priority items to work on
- Follow the existing code patterns and error handling approach
- Add tests for new features
- Update documentation for user-facing changes

### Development Setup
```bash
make help          # See all available commands
make check         # Run tests, formatting, and linting
make build         # Build the binary
```