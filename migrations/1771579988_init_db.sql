-- Enable the required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schema if it doesn’t exist
CREATE SCHEMA IF NOT EXISTS public AUTHORIZATION pg_database_owner;

-- ======================================
-- Table: users
-- ======================================
DROP TABLE IF EXISTS public.users;

CREATE TABLE public.users (
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    username TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_username_key UNIQUE (username)
);

-- ======================================
-- Table: todos
-- ======================================
DROP TABLE IF EXISTS public.todos;
 
CREATE TABLE public.todos (
    id SERIAL NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    assignee TEXT NOT NULL,
    due_date DATE NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL,
    deleted_at TIMESTAMP NULL
);

-- ======================================
-- Function: set_updated_at()
-- ======================================
DROP FUNCTION IF EXISTS set_updated_at();
 
CREATE FUNCTION set_updated_at
RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- ======================================
-- Trigger: trigger_set_updated_at on todos
-- ======================================
DROP TRIGGER IF EXISTS set_trigger_updated_at ON public.todos;

CREATE TRIGGER set_trigger_updated_at
BEFORE UPDATE on public.todos
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

