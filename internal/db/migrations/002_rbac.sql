-- Teams
CREATE TABLE IF NOT EXISTS teams (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Users
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    name          VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    oidc_subject  VARCHAR(255),
    is_admin      BOOLEAN DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login    TIMESTAMP WITH TIME ZONE
);

-- Team memberships
CREATE TABLE IF NOT EXISTS team_members (
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role    VARCHAR(50) NOT NULL DEFAULT 'member',
    PRIMARY KEY (team_id, user_id)
);

-- Cluster permissions (namespace-level RBAC)
CREATE TABLE IF NOT EXISTS cluster_permissions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id    UUID REFERENCES teams(id) ON DELETE CASCADE,
    cluster_id VARCHAR(255) NOT NULL,
    namespace  VARCHAR(255) NOT NULL,
    permission VARCHAR(50)  NOT NULL,
    UNIQUE (team_id, cluster_id, namespace)
);

-- Registered clusters
CREATE TABLE IF NOT EXISTS clusters (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    kubeconfig  BYTEA NOT NULL,
    server_url  VARCHAR(255) NOT NULL,
    version     VARCHAR(50),
    added_by    UUID REFERENCES users(id),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit log
CREATE TABLE IF NOT EXISTS audit_logs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id),
    cluster_id VARCHAR(255),
    namespace  VARCHAR(255),
    resource   VARCHAR(255) NOT NULL,
    action     VARCHAR(50)  NOT NULL,
    details    JSONB,
    ip_address INET,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_cluster ON audit_logs(cluster_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
