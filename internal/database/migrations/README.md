# Database Migrations

This project uses SQL-based migrations instead of GORM AutoMigrate to ensure compatibility with managed database services like Supabase that restrict access to `information_schema`.

## Migration Files

Migration files are located in `internal/database/migrations/` and follow the naming convention:

```
{version}_{description}.up.sql
```

Example: `000001_create_users_table.up.sql`

## How It Works

1. **Migration Tracking**: Applied migrations are tracked in the `schema_migrations` table
2. **Embedded Files**: Migration files are embedded in the binary using Go's `embed` directive
3. **Automatic Execution**: Migrations run automatically on application startup
4. **Idempotent**: Each migration can be safely run multiple times

## Available Migrations

- `000001_create_users_table.up.sql` - Creates users table with indexes
- `000002_create_credentials_table.up.sql` - Creates credentials table with foreign key
- `000003_create_sessions_table.up.sql` - Creates sessions table with indexes

## Adding New Migrations

1. Create a new SQL file in `internal/database/migrations/`
2. Follow the naming convention: `{next_version}_{description}.up.sql`
3. Write your SQL migration using `CREATE TABLE IF NOT EXISTS` syntax
4. The migration will automatically run on next application startup

## Compatibility

This migration system is compatible with:
- PostgreSQL 14+
- Supabase
- AWS RDS
- Google Cloud SQL
- Azure Database for PostgreSQL
- Any managed PostgreSQL service
