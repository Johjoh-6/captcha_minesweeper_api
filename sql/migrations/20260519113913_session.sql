-- +goose Up
CREATE TABLE sessions (
    session_id UUID PRIMARY KEY DEFAULT uuidv4(),
    ip_address VARCHAR(45) NOT NULL, -- IPv4/IPv6
    user_agent TEXT NOT NULL,
    request_count INT DEFAULT 0, -- Track total requests for analytics and avg time
    avg_time_between_requests FLOAT, -- ms (NULL if <2 requests)
    is_bot BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


-- Indexes
CREATE INDEX idx_sessions_ip ON sessions(ip_address);
CREATE INDEX idx_sessions_is_bot ON sessions(is_bot);

-- +goose Down
DROP TABLE sessions;
