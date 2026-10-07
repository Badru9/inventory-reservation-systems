#!/bin/sh
# Dev-only init script: replace the default pg_hba.conf with our trust-everything version
# so the host and other containers can connect with any password.
set -e
cp /etc/postgresql/pg_hba.conf "$PGDATA/pg_hba.conf"
chown postgres:postgres "$PGDATA/pg_hba.conf"
chmod 600 "$PGDATA/pg_hba.conf"
echo "Custom pg_hba.conf installed (trust auth enabled for dev)."
