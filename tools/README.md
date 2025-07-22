# Oil & Gas Inventory - Migration Tools

Modern Go-based tools for converting legacy MDB files and ColdFusion applications to PostgreSQL.

## Quick Start

```bash
# From project root
make migration::setup
make migration::build
make migration::convert FILE=database.mdb COMPANY="Company Name"
```

## Structure

- `cmd/` - Command-line tools (mdb_processor, cf_analyzer, etc.)
- `internal/` - Internal packages (config, processor, mapping, validation)
- `config/` - Configuration files for oil & gas industry standards
- `test/` - Test files and fixtures
- `output/` - Generated conversion outputs

## Development

```bash
# Build tools
go build -o bin/mdb_processor cmd/mdb_processor.go

# Run tests  
go test ./...

# Generate documentation
go doc ./...
```

See the main project Makefile for all available commands.
