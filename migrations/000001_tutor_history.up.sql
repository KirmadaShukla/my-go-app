-- Tutor learning history schema (Postgres)
-- Applied automatically via GORM AutoMigrate on API startup.
-- This file is the readable source of truth for reviewers/ops.

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    gender VARCHAR(20) NOT NULL,
    mother_name VARCHAR(255) NOT NULL,
    father_name VARCHAR(255) NOT NULL,
    mobile_number VARCHAR(20) NOT NULL,
    child_age INTEGER NOT NULL,
    child_class VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS tutor_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    subject VARCHAR(32) NOT NULL,
    language VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    message_count INTEGER NOT NULL DEFAULT 0,
    last_message_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tutor_sessions_user_subject_status
    ON tutor_sessions (user_id, subject, status);

CREATE INDEX IF NOT EXISTS idx_tutor_sessions_user_updated
    ON tutor_sessions (user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS tutor_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES tutor_sessions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL,
    channel VARCHAR(16) NOT NULL DEFAULT 'voice',
    content TEXT NOT NULL,
    sequence BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tutor_messages_session_seq
    ON tutor_messages (session_id, sequence);

CREATE INDEX IF NOT EXISTS idx_tutor_messages_session_created
    ON tutor_messages (session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tutor_messages_user_created
    ON tutor_messages (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS tutor_subject_memories (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    subject VARCHAR(32) NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ux_tutor_subject_memory_user_subject UNIQUE (user_id, subject)
);
