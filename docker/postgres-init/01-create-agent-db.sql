-- Create database for pina-colada application if it doesn't exist
-- This runs automatically when Postgres container first starts

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pina_colada') THEN
        CREATE DATABASE pina_colada;
    END IF;
END
$$;

