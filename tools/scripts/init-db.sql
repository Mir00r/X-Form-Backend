-- Database initialization script for X-Form Backend
-- This script creates the initial database schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    provider VARCHAR(50) DEFAULT 'email', -- 'email', 'google'
    provider_id VARCHAR(255),
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Forms table
CREATE TABLE IF NOT EXISTS forms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schema JSONB NOT NULL, -- Form structure and questions
    settings JSONB DEFAULT '{}', -- Form settings (public, requires login, etc.)
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'published', 'closed'
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Form questions table (normalized from form schema for better querying)
CREATE TABLE IF NOT EXISTS form_questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    question_id VARCHAR(50) NOT NULL, -- ID within the form
    question_type VARCHAR(50) NOT NULL, -- 'text', 'email', 'radio', 'checkbox', etc.
    title TEXT NOT NULL,
    description TEXT,
    required BOOLEAN DEFAULT FALSE,
    options JSONB, -- For radio/checkbox questions
    validation JSONB, -- Validation rules
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Form collaborators table (for team features - post-MVP)
CREATE TABLE IF NOT EXISTS form_collaborators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'viewer', -- 'owner', 'editor', 'viewer'
    invited_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(form_id, user_id)
);

-- File uploads table
CREATE TABLE IF NOT EXISTS file_uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    form_id UUID REFERENCES forms(id) ON DELETE CASCADE,
    original_name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    s3_key TEXT,
    s3_bucket VARCHAR(100),
    status VARCHAR(20) DEFAULT 'processing', -- 'processing', 'ready', 'error'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_forms_user_id ON forms(user_id);
CREATE INDEX IF NOT EXISTS idx_forms_status ON forms(status);
CREATE INDEX IF NOT EXISTS idx_forms_created_at ON forms(created_at);
CREATE INDEX IF NOT EXISTS idx_form_questions_form_id ON form_questions(form_id);
CREATE INDEX IF NOT EXISTS idx_form_questions_order ON form_questions(form_id, order_index);
CREATE INDEX IF NOT EXISTS idx_form_collaborators_form_id ON form_collaborators(form_id);
CREATE INDEX IF NOT EXISTS idx_form_collaborators_user_id ON form_collaborators(user_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_user_id ON file_uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_form_id ON file_uploads(form_id);

-- Updated at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_forms_updated_at BEFORE UPDATE ON forms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (for development)
INSERT INTO users (email, password_hash, first_name, last_name, email_verified)
VALUES (
    'admin@xform.dev',
    '$2b$10$rQJ5X2nX8xGHYm8Kp9YVsOQ8WgH7KXkfZv3PqN2mL4cR5tS6wE7uA', -- password: 'admin123'
    'Admin',
    'User',
    TRUE
) ON CONFLICT (email) DO NOTHING;
