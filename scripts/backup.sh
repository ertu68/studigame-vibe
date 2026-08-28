#!/bin/bash

# Simple backup script for pennergame SQLite database
# This script should be run via cron (e.g., nightly)

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
DB_PATH="${DB_PATH:-./pennergame.db}"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="pennergame_${DATE}.backup"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Create backup
if [ -f "$DB_PATH" ]; then
	echo "Creating backup: $BACKUP_DIR/$BACKUP_NAME"
	cp "$DB_PATH" "$BACKUP_DIR/$BACKUP_NAME"
	echo "Backup completed successfully"
else
	echo "Error: Database file not found at $DB_PATH"
	exit 1
fi

# Cleanup old backups (keep last 30)
echo "Cleaning up old backups (keeping last 30)"
ls -t "$BACKUP_DIR"/pennergame_*.backup | tail -n +31 | xargs -r rm

echo "Backup script completed"
