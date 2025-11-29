# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.0.0] - 2025-11-29

### Added
- Implemented Configuration & Security (Config struct, AES encryption for credentials, init logic).
- Implemented Storage & Metadata (Backup file layout, metadata.json, ListBackups).
- Implemented Utility Layer (Safe shell command execution, Checksums).
- Implemented Database Engines (Interface, MySQL, Postgres, Mongo adapters).
- Implemented CLI & Main Entrypoint (Interactive menu, command flags, wiring).

### Changed

### Fixed
- Improved backup resilience: `BackupAll` now continues even if individual database backups fail (partial success).
- Added retry logic (3 retries) for PostgreSQL backups to handle transient network errors (e.g., SSL SYSCALL error).
- Metadata now records per-file status ("success" or "failed") and error messages.
- Resolved MongoDB authentication error by passing password explicitly via flag instead of unreliable stdin piping.
- Added retry logic (3 retries) for MongoDB backups to handle transient network issues (e.g., server selection timeout).

### Removed

### Documentation
- Updated README with usage instructions and fixed build errors.
- Added initial README.md.
- Removed reference to unnecessary configuration from README.

### Chores
- Initialize project structure and dependencies.
- Added dbmigrate to .gitignore.
- Ignore build artifact.
- Initialize go.mod for mydbportal.com/dbmigrate.
