# Parts Examples

This directory contains practical examples demonstrating how to use Parts with different file types and comment styles.

## Directory Structure

- `ssh-config/` - SSH configuration management (hash comments)
- `hosts/` - System hosts file management (hash comments)
- `javascript/` - JavaScript configuration merging (slash comments)
- `css/` - CSS style merging (block comments)
- `sql/` - SQL script merging (dash comments)  
- `python/` - Python configuration (auto-detection)
- `batch/` - Windows batch file merging (rem comments)

## Running Examples

Each directory contains:
- A base configuration file
- A `partials/` directory with partial files
- A `run-example.sh` script to demonstrate usage

### Quick Start

```bash
# Run all examples
make examples

# Or run individual examples
cd examples/ssh-config && ./run-example.sh
cd examples/hosts && ./run-example.sh
cd examples/javascript && ./run-example.sh
# etc.
```

## Comment Styles Supported

| Style | File Types | Example |
|-------|------------|---------|
| `#` | Shell, Python, YAML, SSH config, hosts | `parts ~/.ssh/config ~/.ssh/config.d "#"` |
| `//` | JavaScript, Go, C++, Java | `parts app.js ./js-partials "//"` |
| `--` | SQL, Lua, Haskell | `parts schema.sql ./sql-partials "auto"` |
| `/*` | CSS, C | `parts styles.css ./css-partials "/*"` |
| `;` | INI files, Lisp | `parts config.ini ./ini-partials ";"` |
| `%` | LaTeX, Erlang | `parts document.tex ./tex-partials "%"` |
| `<!--` | HTML, XML | `parts index.html ./html-partials "<!--"` |
| `'` | VB, some config files | `parts config.vb ./vb-partials "'"` |
| `rem` | Batch files | `parts script.bat ./bat-partials "rem"` |
| `auto` | Auto-detect from extension | `parts config.py ./partials "auto"` |

## Custom Comment Characters

You can also use any custom string as a comment character:

```bash
parts myfile.txt ./partials "###"
parts config.custom ./partials ">>"
```