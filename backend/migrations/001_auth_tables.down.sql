-- Auth Module: Rollback Schema
-- This migration drops all tables created by 001_auth_tables.up.sql

-- Drop triggers first
DROP TRIGGER IF EXISTS update_user_tenant_roles_updated_at ON user_tenant_roles;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS auth_events;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_tenant_roles;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS users;
