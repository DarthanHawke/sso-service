CREATE TABLE entities (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL, -- 'user', 'payment', 'session'
    attributes JSONB
);

CREATE TABLE relations (
    id UUID PRIMARY KEY,
    source_id UUID NOT NULL REFERENCES entities(id),
    target_id UUID NOT NULL REFERENCES entities(id),
    relation_type VARCHAR(50) NOT NULL, -- 'owner', 'member', 'admin', 'support'
    attributes JSONB
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE permission_assignments (
    relation_type VARCHAR(50) NOT NULL,
    permission_id UUID NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (relation_type, permission_id)
);