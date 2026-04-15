-- Migration: 001_create_users
-- Creates the core users table linking OAuth identities to tenant pod instances.

CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    oauth_id   VARCHAR(255) UNIQUE NOT NULL,
    tenant_id  VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
