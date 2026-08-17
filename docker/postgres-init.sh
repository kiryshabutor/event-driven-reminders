#!/bin/sh

set -eu

for migration in \
  /project-migrations/auth/*.up.sql \
  /project-migrations/reminder/*.up.sql \
  /project-migrations/analytics/*.up.sql
do
  [ -f "$migration" ] || continue
  echo "Applying migration: $migration"
  psql -v ON_ERROR_STOP=1 \
    --username "$POSTGRES_USER" \
    --dbname "$POSTGRES_DB" \
    --file "$migration"
done
