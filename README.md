# Parts

[![CI/CD Pipeline](https://github.com/cageis/parts/actions/workflows/ci.yml/badge.svg)](https://github.com/cageis/parts/actions/workflows/ci.yml)
[![Code Quality](https://github.com/cageis/parts/actions/workflows/quality.yml/badge.svg)](https://github.com/cageis/parts/actions/workflows/quality.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/cageis/parts)](https://goreportcard.com/report/github.com/cageis/parts)
[![codecov](https://codecov.io/gh/cageis/parts/branch/main/graph/badge.svg)](https://codecov.io/gh/cageis/parts)

A Go utility for merging partial configuration files into an aggregate file. Perfect for managing configurations that don't support native includes - like SSH configs, `/etc/hosts`, and other system files.

## Features

- **Idempotent Operations**: Run multiple times without causing file bloat or whitespace growth
- **Automatic Section Management**: Uses comment markers to manage the merged section
- **Original Content Preservation**: Your existing configuration remains untouched
- **Dynamic Updates**: Add/remove partial files and re-run to update the aggregate
- **Source Tracking**: Each partial section includes a comment showing the source file path for easy identification
- **Dry-Run Mode**: Preview changes without modifying files using `--dry-run` or `-n`
- **Remove Mode**: Clean removal of partials sections using `--remove` or `-r`
- **Language-Aware Comment Styles**: Automatically uses the correct comment syntax for your file type to ensure markers don't break functionality

## Installation

### From Source

```bash
git clone https://github.com/cageis/parts.git
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

### Common Configuration Examples

**SSH Configuration:**
```bash
# Build: Merge SSH config partials
parts ~/.ssh/config ~/.ssh/config.d "#"

# Remove: Remove partials section from SSH config
parts --remove ~/.ssh/config "#"

# Preview changes without modifying files (dry-run)
parts --dry-run ~/.ssh/config ~/.ssh/config.d "#"
```

**System Hosts File:**
```bash
# Build: Merge host file partials (useful for managing dev/staging/prod hosts)
sudo parts /etc/hosts /etc/hosts.d "#"

# Remove: Clean up partials section
sudo parts --remove /etc/hosts "#"

# Preview changes first (recommended for system files)
sudo parts --dry-run /etc/hosts /etc/hosts.d "#"
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

**Important:** The comment style parameter isn't cosmetic - it ensures the section markers use valid comment syntax for your file type. This prevents syntax errors, broken parsing, or execution issues that would occur if the wrong comment style was used.

## Examples

### SSH Configuration Management

#### Directory Structure
```
~/.ssh/
├── config                    # Main SSH config
└── config.d/                # Partials directory
    ├── work-servers         # Work-related SSH hosts
    ├── personal-servers     # Personal SSH hosts
    └── development          # Development environment hosts
```

#### Partial File Example (`~/.ssh/config.d/work-servers`)
```
Host work-db
    HostName db.company.com
    User dbadmin
    Port 2222

Host work-app
    HostName app.company.com  
    User deploy
```

#### Result in `~/.ssh/config`
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

### System Hosts File Management

Perfect for managing development, staging, and production host mappings without editing the main system file directly.

#### Directory Structure
```
/etc/
├── hosts                     # Main system hosts file
└── hosts.d/                  # Partials directory
    ├── development          # Dev environment hosts
    ├── staging              # Staging hosts
    ├── production           # Production hosts
    └── local-services       # Local service mappings
```

#### Partial File Example (`/etc/hosts.d/development`)
```
# Development Environment
127.0.0.1    api.dev.company.com
127.0.0.1    app.dev.company.com
192.168.1.100 db.dev.company.com
192.168.1.101 cache.dev.company.com
```

#### Result in `/etc/hosts`
```
# Standard system entries
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost

# ============================
# PARTIALS>>>>>
# ============================
# Source: /etc/hosts.d/development
# Development Environment
127.0.0.1    api.dev.company.com
127.0.0.1    app.dev.company.com
192.168.1.100 db.dev.company.com
192.168.1.101 cache.dev.company.com
# Source: /etc/hosts.d/production
# Production Environment  
10.0.1.50    api.company.com
10.0.1.51    app.company.com
# ============================
# PARTIALS<<<<<
# ============================
```

### Comment Styles by File Type

Parts automatically chooses the right comment style for your file type to ensure the section markers are valid syntax. The markers need to be proper comments so they don't break syntax highlighting, parsing, or execution.

| File Type | Comment Style | Example Usage | Marker Format |
|-----------|---------------|---------------|---------------|
| SSH Config, Hosts, Shell, Python | `#` | `parts ~/.ssh/config ~/.ssh/config.d "#"` | `# PARTIALS>>>>>` |
| JavaScript, Go, C++, Java | `//` | `parts app.js ./js-partials "//"` | `// PARTIALS>>>>>` |
| CSS, C-style blocks | `/*` | `parts styles.css ./css-partials "/*"` | `/* PARTIALS>>>>> */` |
| SQL, Lua, Haskell | `--` | `parts schema.sql ./sql-partials "--"` | `-- PARTIALS>>>>>` |
| HTML, XML | `<!--` | `parts index.html ./html-partials "<!--"` | `<!-- PARTIALS>>>>> -->` |
| AutoLISP, Emacs Lisp | `;` | `parts config.el ./lisp-configs ";"` | `; PARTIALS>>>>>` |
| MATLAB, TeX | `%` | `parts analysis.m ./matlab-scripts "%"` | `% PARTIALS>>>>>` |
| Auto-detection | `auto` | `parts config.py ./python-configs "auto"` | *Detects from extension* |

**Why different comment styles?** Each programming language and configuration format has its own comment syntax. Using the wrong comment style would create syntax errors or break your file's functionality.

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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

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