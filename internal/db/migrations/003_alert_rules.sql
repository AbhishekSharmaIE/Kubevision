CREATE TABLE IF NOT EXISTS alert_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    query       TEXT NOT NULL,
    threshold   DOUBLE PRECISION NOT NULL,
    comparator  VARCHAR(10) NOT NULL,
    duration    VARCHAR(50) NOT NULL DEFAULT '5m',
    severity    VARCHAR(50) NOT NULL,
    channels    JSONB DEFAULT '[]',
    cluster_id  VARCHAR(255) NOT NULL,
    created_by  UUID REFERENCES users(id),
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
