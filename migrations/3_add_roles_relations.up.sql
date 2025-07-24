CREATE TABLE relations (
    id UUID PRIMARY KEY,
    source_id UUID NOT NULL REFERENCES entities(id),
    target_id UUID NOT NULL REFERENCES entities(id),
    relation_type VARCHAR(50) NOT NULL, -- 'owner', 'member', 'admin', 'support'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    attributes JSONB
);

CREATE TABLE permission_assignments (
    relation_type VARCHAR(50) NOT NULL,
    permission_id UUID NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (relation_type, permission_id)
);