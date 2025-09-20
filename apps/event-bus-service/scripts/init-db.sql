-- Event Bus Service - Database Initialization Script
-- This script sets up the PostgreSQL database for the Event Bus Service
-- with tables for forms, responses, and analytics to demonstrate CDC

-- Create database user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'eventbus') THEN
        CREATE ROLE eventbus LOGIN PASSWORD 'eventbus_password';
    END IF;
END
$$;

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE eventbus TO eventbus;
GRANT ALL PRIVILEGES ON SCHEMA public TO eventbus;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO eventbus;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO eventbus;

-- Enable logical replication for Debezium
ALTER SYSTEM SET wal_level = logical;
ALTER SYSTEM SET max_wal_senders = 4;
ALTER SYSTEM SET max_replication_slots = 4;

-- Create tables for demonstration
-- Forms table
CREATE TABLE IF NOT EXISTS public.forms (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB NOT NULL,
    settings JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'draft',
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

-- Responses table
CREATE TABLE IF NOT EXISTS public.responses (
    id SERIAL PRIMARY KEY,
    form_id INTEGER NOT NULL REFERENCES public.forms(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'submitted',
    submitted_by INTEGER,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

-- Analytics table
CREATE TABLE IF NOT EXISTS public.analytics (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    user_id INTEGER,
    session_id VARCHAR(255),
    data JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_forms_created_by ON public.forms(created_by);
CREATE INDEX IF NOT EXISTS idx_forms_status ON public.forms(status);
CREATE INDEX IF NOT EXISTS idx_forms_created_at ON public.forms(created_at);

CREATE INDEX IF NOT EXISTS idx_responses_form_id ON public.responses(form_id);
CREATE INDEX IF NOT EXISTS idx_responses_submitted_by ON public.responses(submitted_by);
CREATE INDEX IF NOT EXISTS idx_responses_submitted_at ON public.responses(submitted_at);
CREATE INDEX IF NOT EXISTS idx_responses_status ON public.responses(status);

CREATE INDEX IF NOT EXISTS idx_analytics_event_type ON public.analytics(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_entity ON public.analytics(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_analytics_user_id ON public.analytics(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_created_at ON public.analytics(created_at);

-- Create GIN indexes for JSONB columns
CREATE INDEX IF NOT EXISTS idx_forms_schema_gin ON public.forms USING GIN(schema);
CREATE INDEX IF NOT EXISTS idx_forms_settings_gin ON public.forms USING GIN(settings);
CREATE INDEX IF NOT EXISTS idx_responses_data_gin ON public.responses USING GIN(data);
CREATE INDEX IF NOT EXISTS idx_responses_metadata_gin ON public.responses USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_analytics_data_gin ON public.analytics USING GIN(data);
CREATE INDEX IF NOT EXISTS idx_analytics_metadata_gin ON public.analytics USING GIN(metadata);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_forms_updated_at ON public.forms;
CREATE TRIGGER update_forms_updated_at
    BEFORE UPDATE ON public.forms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_responses_updated_at ON public.responses;
CREATE TRIGGER update_responses_updated_at
    BEFORE UPDATE ON public.responses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create publication for logical replication (required for Debezium)
DROP PUBLICATION IF EXISTS eventbus_publication;
CREATE PUBLICATION eventbus_publication FOR TABLE public.forms, public.responses, public.analytics;

-- Grant replication permission to eventbus user
ALTER USER eventbus REPLICATION;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO eventbus;

-- Insert sample data for testing
INSERT INTO public.forms (title, description, schema, settings, created_by) VALUES
('User Feedback Survey', 'Collect user feedback about our application', 
 '{"fields": [{"name": "rating", "type": "number", "label": "Rating", "required": true}, {"name": "comments", "type": "textarea", "label": "Comments"}]}',
 '{"allow_anonymous": true, "collect_ip": true}', 1),
('Customer Information Form', 'Collect customer contact information',
 '{"fields": [{"name": "name", "type": "text", "label": "Full Name", "required": true}, {"name": "email", "type": "email", "label": "Email", "required": true}, {"name": "phone", "type": "tel", "label": "Phone Number"}]}',
 '{"allow_anonymous": false, "require_login": true}', 1),
('Product Feature Request', 'Submit requests for new product features',
 '{"fields": [{"name": "feature_title", "type": "text", "label": "Feature Title", "required": true}, {"name": "description", "type": "textarea", "label": "Description", "required": true}, {"name": "priority", "type": "select", "label": "Priority", "options": ["Low", "Medium", "High"]}]}',
 '{"allow_anonymous": false, "notify_admin": true}', 2);

-- Insert sample responses
INSERT INTO public.responses (form_id, data, submitted_by, ip_address) VALUES
(1, '{"rating": 5, "comments": "Great application, very user-friendly!"}', 101, '192.168.1.10'),
(1, '{"rating": 4, "comments": "Good overall, but could use some improvements"}', 102, '192.168.1.11'),
(2, '{"name": "John Doe", "email": "john.doe@example.com", "phone": "+1234567890"}', 103, '192.168.1.12'),
(3, '{"feature_title": "Dark Mode", "description": "Add dark mode support for better user experience", "priority": "High"}', 104, '192.168.1.13');

-- Insert sample analytics data
INSERT INTO public.analytics (event_type, entity_type, entity_id, user_id, session_id, data, ip_address) VALUES
('form_view', 'form', 1, 101, 'sess_123', '{"referrer": "https://example.com", "browser": "Chrome"}', '192.168.1.10'),
('form_submit', 'form', 1, 101, 'sess_123', '{"completion_time": 120, "errors": 0}', '192.168.1.10'),
('form_view', 'form', 2, 103, 'sess_456', '{"referrer": "direct", "browser": "Firefox"}', '192.168.1.12'),
('form_submit', 'form', 2, 103, 'sess_456', '{"completion_time": 95, "errors": 1}', '192.168.1.12'),
('response_view', 'response', 1, 1, 'sess_admin', '{"admin_action": "review"}', '192.168.1.1');

-- Create function to generate test data
CREATE OR REPLACE FUNCTION generate_test_data(num_forms INTEGER DEFAULT 10, num_responses INTEGER DEFAULT 50)
RETURNS VOID AS $$
DECLARE
    form_id INTEGER;
    i INTEGER;
BEGIN
    -- Generate test forms
    FOR i IN 1..num_forms LOOP
        INSERT INTO public.forms (title, description, schema, created_by)
        VALUES (
            'Test Form ' || i,
            'This is a test form generated for CDC testing - ' || i,
            '{"fields": [{"name": "field_' || i || '", "type": "text", "label": "Test Field ' || i || '"}]}',
            (RANDOM() * 10 + 1)::INTEGER
        ) RETURNING id INTO form_id;
        
        -- Generate responses for this form
        FOR j IN 1..(num_responses / num_forms) LOOP
            INSERT INTO public.responses (form_id, data, submitted_by, ip_address)
            VALUES (
                form_id,
                ('{"field_' || i || '": "Test response ' || j || ' for form ' || i || '"}')::JSONB,
                (RANDOM() * 100 + 1)::INTEGER,
                ('192.168.1.' || (RANDOM() * 254 + 1)::INTEGER)::INET
            );
            
            -- Generate analytics for this response
            INSERT INTO public.analytics (event_type, entity_type, entity_id, user_id, data)
            VALUES (
                CASE (RANDOM() * 3)::INTEGER
                    WHEN 0 THEN 'form_view'
                    WHEN 1 THEN 'form_submit'
                    ELSE 'response_view'
                END,
                'form',
                form_id,
                (RANDOM() * 100 + 1)::INTEGER,
                ('{"test_data": true, "iteration": ' || j || '}')::JSONB
            );
        END LOOP;
    END LOOP;
    
    RAISE NOTICE 'Generated % test forms with % responses each', num_forms, (num_responses / num_forms);
END;
$$ LANGUAGE plpgsql;

-- Grant execute permission on the function
GRANT EXECUTE ON FUNCTION generate_test_data TO eventbus;

-- Create view for CDC monitoring
CREATE OR REPLACE VIEW cdc_activity_summary AS
SELECT 
    'forms' as table_name,
    COUNT(*) as total_records,
    MAX(updated_at) as last_updated
FROM public.forms
UNION ALL
SELECT 
    'responses' as table_name,
    COUNT(*) as total_records,
    MAX(updated_at) as last_updated
FROM public.responses
UNION ALL
SELECT 
    'analytics' as table_name,
    COUNT(*) as total_records,
    MAX(created_at) as last_updated
FROM public.analytics;

-- Grant select on the view
GRANT SELECT ON cdc_activity_summary TO eventbus;

-- Print completion message
DO $$
BEGIN
    RAISE NOTICE 'Event Bus Service database initialization completed successfully!';
    RAISE NOTICE 'Tables created: forms, responses, analytics';
    RAISE NOTICE 'Sample data inserted for testing CDC functionality';
    RAISE NOTICE 'Logical replication configured for Debezium';
    RAISE NOTICE 'Use function generate_test_data() to create more test data';
END
$$;
