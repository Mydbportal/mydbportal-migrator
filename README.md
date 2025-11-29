# DB-Migrate-Go

A cross-database migration tool written in Go that supports backing up, storing, and restoring databases for MySQL, PostgreSQL, and MongoDB.

## Features

- **Backup**: Backup all databases or a specific one from a source server.
- **Restore**: Restore a backup to a target server.
- **Secure**: Encrypts stored credentials using AES-256-GCM.
- **Interactive CLI**: User-friendly interactive menu.
- **Metadata**: JSON metadata with checksums for every backup.

## Installation

1. Clone the repository.
2. Build the binary:
   ```bash
   go build -o dbmigrate cmd/dbmigrate/main.go
   ```
3. Ensure you have the native tools installed on your system/path:
   - `mysql`, `mysqldump` (for MySQL)
   - `psql`, `pg_dump` (for PostgreSQL)
   - `mongosh`, `mongodump`, `mongorestore` (for MongoDB)

## Usage

### Interactive Mode (Recommended)

Run the tool without arguments or with `interactive` command:

```bash
./dbmigrate interactive
```

Follow the numbered menu to Add Source, Backup, List, or Restore.

### Command Line Arguments

#### 1. Initialize (Add Source)
```bash
./dbmigrate init
```
Follow prompts to add source details.

#### 2. Backup
Backup all databases from a source (use ID from init):
```bash
./dbmigrate backup --source my-mysql-server
```
Backup a specific database:
```bash
./dbmigrate backup --source my-mysql-server --db my_database
```

#### 3. List Backups
```bash
./dbmigrate list
```

#### 4. Restore
Restore a backup file to a target server:
```bash
./dbmigrate restore --backup backups/mysql/source-127.0.0.1_.../db1_...sql.gz --target my-target-server
```

## Configuration

Configuration is stored in `~/.dbmigrate.json` (or `.gemini/config.json` in this environment). Credentials are encrypted.

## Project Structure

- `cmd/dbmigrate`: Main entry point.
- `internal/engine`: Database specific implementations.
- `internal/storage`: Backup file and metadata management.
- `internal/config`: Configuration and credential management.
- `internal/cli`: CLI logic and interactive menu.