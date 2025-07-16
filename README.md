# Oil & Gas Inventory System

A modern web application for managing oil & gas inventory, migrated from ColdFusion/Access to Go/PostgreSQL.

## Project Phases

This project is organized into phases, each building upon the previous one:

### 📋 [Phase 1: MDB Migration](docs/README_PHASE1.md)
**Current Phase** - Migrate Access MDB database to normalized CSV files
- ✅ Fixes EOF errors in MDB column analysis
- ✅ Extracts and normalizes database tables  
- ✅ Generates analysis files for development
- ✅ Prepares PostgreSQL-compatible data files

**Quick Start:**
```bash
./setup_phase1.sh    # Complete Phase 1 setup and migration
```

### 🚀 [Phase 2: Local Development Setup](docs/README_PHASE2.md)
Set up local development environment with Docker
- 🐘 PostgreSQL database with Docker
- 🗄️ PgAdmin for database management
- 📊 Import normalized CSV data from Phase 1
- ⚡ Hot-reload development servers

**Quick Start:**
```bash
make setup           # After Phase 1 completes
make dev
```

### 🏗️ [Phase 3: Production Deployment](docs/README_PHASE3.md)
Deploy to production environment
- 🚀 Vultr VPS deployment
- 🔐 SSL/TLS configuration
- 📈 Performance monitoring
- 🔄 CI/CD pipeline

### 🔧 [Phase 4: Advanced Features](docs/README_PHASE4.md)
Advanced functionality and optimization
- 📱 Mobile responsiveness
- 🔍 Advanced search and filtering
- 📊 Analytics and reporting
- 🔌 API integrations

## Current Status

### ✅ Phase 1: MDB Migration
Run `./setup_phase1.sh` to complete Phase 1 migration.

Generated files will include:
- `database/analysis/mdb_column_analysis.txt` - Column mapping analysis
- `database/analysis/table_counts.txt` - Table record counts
- `database/analysis/table_list.txt` - Table inventory  
- `database/data/clean/` - Normalized CSV files ready for import
- `database/schema/mdb_schema.sql` - Database schema

### 🚧 Phase 2: Local Development Setup
Available after Phase 1 completes successfully.

## Quick Navigation

### Documentation
- 📋 [Phase 1 Documentation](docs/README_PHASE1.md) - MDB Migration
- 🚀 [Phase 2 Documentation](docs/README_PHASE2.md) - Local Development  
- 📖 [Technical Architecture](docs/ARCHITECTURE.md) - System design
- 🔧 [Development Guide](docs/DEVELOPMENT.md) - Developer workflow

### Scripts
- 📄 [`./setup_phase1.sh`](setup_phase1.sh) - Phase 1 complete setup
- 📄 [`scripts/phase1_mdb_migration.sh`](scripts/phase1_mdb_migration.sh) - Reusable migration script
- 📄 [`scripts/`](scripts/) - All utility scripts

### Configuration
- 🐳 [`docker-compose.yml`](docker-compose.yml) - Local development services
- ⚙️ [`Makefile`](Makefile) - Development commands
- 🔧 [`.env.local`](.env.local) - Local environment configuration

## Technology Stack

### Current (Phase 1)
- **Data Processing**: Go with enhanced MDB tools
- **Migration**: Access MDB → PostgreSQL via CSV
- **Analysis**: Automated column mapping and validation

### Target (Phase 2+)
- **Backend**: Go with Gin framework
- **Frontend**: Vue.js 3 with TypeScript
- **Database**: PostgreSQL 15
- **Infrastructure**: Docker + Docker Compose
- **Development**: Hot-reload, automated testing

## Data Security

### ✅ Safe for Version Control
- Mock/example data in development seeds
- Documentation and source code
- Configuration templates
- Database schema (structure only)

### ❌ Protected by .gitignore
- Real customer data and business information
- Production credentials and API keys
- Database files (.mdb, backups, exports)
- Environment files with real credentials

## Migration Notes

This system was migrated from ColdFusion/Access to Go/PostgreSQL:

1. **Schema Conversion**: Access → PostgreSQL with type mapping
2. **Data Migration**: CSV export with date/case normalization  
3. **Query Analysis**: ColdFusion queries analyzed for optimization
4. **Grade Validation**: Oil & gas industry grades (J55, JZ55, L80, N80, P105, P110)

## Getting Started

### Prerequisites
- Go 1.19+ (`go version`)
- Docker & Docker Compose (`docker --version`)
- mdb-tools (`mdb-tables --version`) - for Phase 1 only

### Installation
```bash
# macOS
brew install go docker mdbtools

# Ubuntu/Debian  
sudo apt-get install golang-go docker.io docker-compose mdb-tools

# Start with Phase 1
./setup_phase1.sh
```

## Project Structure

```
oil-gas-inventory/
├── README.md                    # ← This file (project overview)
├── setup_phase1.sh              # ← Phase 1 complete setup
├── docs/                        # ← All documentation
│   ├── README_PHASE1.md         # Phase 1 guide
│   ├── README_PHASE2.md         # Phase 2 guide  
│   └── ...
├── scripts/                     # ← All utility scripts
│   ├── phase1_mdb_migration.sh  # Reusable migration script
│   └── ...
├── database/                    # ← Generated data and analysis
│   ├── analysis/                # Analysis files for development
│   ├── data/clean/              # Normalized CSV files
│   └── schema/                  # Database schema files
├── backend/                     # ← Go backend application
├── frontend/                    # ← Vue.js frontend application
├── db_prep/                     # ← Original MDB file location
├── docker-compose.yml           # ← Local development services
├── Makefile                     # ← Development commands
└── .env.local                   # ← Local environment template
```

## Contributing

1. Follow Go and Vue.js best practices
2. Use conventional commits
3. Add tests for new features
4. Update documentation for new phases
5. **NEVER** commit real customer data or production credentials

## Support

### Current Phase Issues
- **Phase 1**: See [Phase 1 Documentation](docs/README_PHASE1.md#troubleshooting)
- **General**: Check [Development Guide](docs/DEVELOPMENT.md)

### Common Commands
```bash
# Phase 1: MDB Migration
./setup_phase1.sh

# Phase 2: Local Development (after Phase 1)
make setup
make dev

# Check status
make status

# View logs
docker-compose logs
```

## License

[Your License Here]

---

**🚀 Ready to start? Run `./setup_phase1.sh` to begin Phase 1 migration!**
