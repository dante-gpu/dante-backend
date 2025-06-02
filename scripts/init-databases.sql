-- DanteGPU Platform Database Initialization
-- This script creates all required databases and users for the platform

-- Create databases for each service
CREATE DATABASE dante_auth;
CREATE DATABASE dante_billing;
CREATE DATABASE dante_registry;
CREATE DATABASE dante_scheduler;
CREATE DATABASE dante_storage;

-- Grant privileges to the main user for all databases
GRANT ALL PRIVILEGES ON DATABASE dante_auth TO dante_user;
GRANT ALL PRIVILEGES ON DATABASE dante_billing TO dante_user;
GRANT ALL PRIVILEGES ON DATABASE dante_registry TO dante_user;
GRANT ALL PRIVILEGES ON DATABASE dante_scheduler TO dante_user;
GRANT ALL PRIVILEGES ON DATABASE dante_storage TO dante_user;

-- Connect to each database and create necessary extensions
\c dante_auth;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c dante_billing;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c dante_registry;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c dante_scheduler;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c dante_storage;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Switch back to the main database
\c dante_platform;

-- Log successful initialization
SELECT 'DanteGPU Platform databases initialized successfully!' as status; 