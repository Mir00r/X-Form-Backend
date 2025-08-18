#!/bin/bash
set -e

# PostgreSQL Multi-Database Initialization Script
# This script creates multiple databases for the X-Form Backend application

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create auth database if it doesn't exist
    SELECT 'CREATE DATABASE xform_auth'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'xform_auth')\gexec

    -- Create forms database if it doesn't exist  
    SELECT 'CREATE DATABASE xform_forms'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'xform_forms')\gexec

    -- Grant privileges to the user
    GRANT ALL PRIVILEGES ON DATABASE xform_auth TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON DATABASE xform_forms TO "$POSTGRES_USER";
    
    -- Switch to auth database and create extensions if needed
    \c xform_auth;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    
    -- Switch to forms database and create extensions if needed
    \c xform_forms;
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOSQL

echo "✅ Multiple databases initialized successfully!"
