#!/bin/bash

# Source database configuration from .env.db file
if [ -f .env.db ]; then
    source .env.db
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') ERROR: .env.db file not found." >> ./database/logs/backup.log
    exit 1
fi

# Check for required environment variables
required_vars=("DB_HOST" "DB_PORT" "DB_NAME" "DB_USER" "DB_PASSWORD" "BACKUP_DIR")

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: $var is not set. Please check your .env file." >> ./database/logs/backup.log
        exit 1
    fi
done

# Find the most recent backup file
MOST_RECENT_BACKUP=$(ls -1t "$BACKUP_DIR" | grep '^backup_' | head -n 1)

if [ -z "$MOST_RECENT_BACKUP" ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') ERROR: No backup files found in $BACKUP_DIR." >> ./database/logs/backup.log
    exit 1
fi

# Full path to the most recent backup file
BACKUP_FILE="$BACKUP_DIR/$MOST_RECENT_BACKUP"

# Restore database with password
if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" < "$BACKUP_FILE"; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') INFO: Database restore completed successfully from $MOST_RECENT_BACKUP." >> ./database/logs/backup.log
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') ERROR: Database restore failed. Please check for errors." >> ./database/logs/backup.log
    exit 1
fi
