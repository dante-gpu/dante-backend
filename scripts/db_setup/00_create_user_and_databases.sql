-- Consolidated Initialization Script for DanteGPU Platform PostgreSQL Databases

-- Ensure the script can be run multiple times without error (idempotency)

-- 1. Create the primary user if it doesn't exist and set password
-- Note: CREATE USER IF NOT EXISTS is not standard SQL for all Postgres versions directly in a script.
-- We'll try to create and handle potential errors, or assume a superuser runs this.
-- For Docker entrypoint, it usually runs as postgres user.

DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'dante_user') THEN

      CREATE ROLE dante_user LOGIN PASSWORD 'dante_password';
   ELSE
      ALTER ROLE dante_user WITH LOGIN PASSWORD 'dante_password';
   END IF;
END
$do$;

-- 2. Create databases if they don't exist and assign ownership

-- Function to create database if not exists and grant ownership
CREATE OR REPLACE FUNCTION create_database_if_not_exists(dbname text, owner_role text) RETURNS void AS $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = dbname) THEN
        EXECUTE format('CREATE DATABASE %I', dbname);
    END IF;
    EXECUTE format('ALTER DATABASE %I OWNER TO %I', dbname, owner_role);
END;
$$ LANGUAGE plpgsql;

-- Create databases for each service
SELECT create_database_if_not_exists('dante_auth', 'dante_user');
SELECT create_database_if_not_exists('dante_billing', 'dante_user');
SELECT create_database_if_not_exists('dante_registry', 'dante_user');
SELECT create_database_if_not_exists('dante_scheduler', 'dante_user');
SELECT create_database_if_not_exists('dante_storage', 'dante_user');

-- (Optional) Grant all privileges on these databases to the user if ownership is not enough
-- GRANT ALL PRIVILEGES ON DATABASE dante_auth TO dante_user;
-- GRANT ALL PRIVILEGES ON DATABASE dante_billing TO dante_user;
-- GRANT ALL PRIVILEGES ON DATABASE dante_registry TO dante_user;
-- GRANT ALL PRIVILEGES ON DATABASE dante_scheduler TO dante_user;
-- GRANT ALL PRIVILEGES ON DATABASE dante_storage TO dante_user;

-- 3. (Optional) Create extensions if needed in specific databases
-- Example for dante_auth if it needs pgcrypto or uuid-ossp
-- \c dante_auth
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

SELECT 'DanteGPU databases and user initialization complete.' AS status; 