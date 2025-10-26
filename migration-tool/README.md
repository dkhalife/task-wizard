# Task Wizard Migration Tool

This tool migrates data from a SQLite database to a MariaDB database while maintaining data integrity and referential constraints.

## Features

- **Read-only SQLite access**: Opens the source SQLite database in read-only mode with immutable flag to ensure data safety
- **Transaction-based migration**: All data is migrated within a single MariaDB transaction, ensuring atomicity
- **Dependency-aware ordering**: Migrates tables in the correct order to satisfy foreign key constraints
- **Uses existing models**: Leverages the same data models as the Task Wizard API server

## Prerequisites

- Go 1.25 or later
- Source SQLite database file
- Target MariaDB database (must already exist and be accessible)

## Building

```bash
cd migration-tool
go build -o migrate
```

## Usage

```bash
./migrate \
  --sqlite /path/to/task-wizard.db \
  --maria-host localhost \
  --maria-port 3306 \
  --maria-db taskwizard \
  --maria-user taskuser \
  --maria-pass taskpass
```

### Command-Line Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--sqlite` | Yes | - | Path to the SQLite database file (read-only) |
| `--maria-host` | No | `localhost` | MariaDB host |
| `--maria-port` | No | `3306` | MariaDB port |
| `--maria-db` | Yes | - | MariaDB database name |
| `--maria-user` | Yes | - | MariaDB username |
| `--maria-pass` | No | - | MariaDB password |

## Migration Order

The tool migrates data in the following order to ensure foreign key constraints are satisfied:

1. Users
2. Labels
3. Tasks
4. Task History
5. Notifications
6. App Tokens
7. User Password Reset Tokens
8. Notification Settings
9. Task-Label Associations (join table)

## Important Notes

- **No schema migration**: This tool only copies data. The target MariaDB database must already have the correct schema created (use the Task Wizard API server with `database.migration: true` to create the schema).
- **Read-only source**: The SQLite database is opened in read-only mode to prevent any accidental modifications.
- **Transaction safety**: All operations are wrapped in a transaction. If any error occurs, the entire migration is rolled back.
- **No duplicate checks**: The tool assumes the target database is empty. Running it multiple times will result in duplicate data or errors.

## Example Workflow

1. Create the target MariaDB database:
   ```sql
   CREATE DATABASE taskwizard CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'taskuser'@'localhost' IDENTIFIED BY 'taskpass';
   GRANT ALL PRIVILEGES ON taskwizard.* TO 'taskuser'@'localhost';
   FLUSH PRIVILEGES;
   ```

2. Run the Task Wizard API server once with the MariaDB configuration to create the schema:
   ```yaml
   database:
     type: mysql
     host: localhost
     port: 3306
     database: taskwizard
     username: taskuser
     password: taskpass
     migration: true
   ```

3. Stop the API server.

4. Run the migration tool:
   ```bash
   ./migrate \
     --sqlite /config/task-wizard.db \
     --maria-host localhost \
     --maria-db taskwizard \
     --maria-user taskuser \
     --maria-pass taskpass
   ```

5. Verify the migration was successful.

6. Update your API server configuration to use MariaDB.

7. Start the API server with the new configuration.

## Error Handling

If the migration fails:
- All changes are automatically rolled back due to the transaction
- An error message will be displayed indicating which step failed
- The source SQLite database remains unmodified (read-only mode)
- You can fix the issue and re-run the migration

## Security Considerations

- The tool does not validate or sanitize data; it assumes the source data is valid
- Connection credentials are passed via command-line flags (consider using environment variables in production)
- The SQLite database is opened in immutable read-only mode to prevent any modifications
