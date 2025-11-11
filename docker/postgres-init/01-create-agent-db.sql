-- Create databases for pina-colada application
-- This runs automatically when Postgres container first starts
-- Note: CREATE DATABASE cannot be executed from a function/DO block
-- These will error if databases already exist, but that's fine for init scripts

-- Create development database for agent/client
CREATE DATABASE development;

-- Create langfuse database (Langfuse will create it if missing, but this avoids startup errors)
CREATE DATABASE langfuse;

