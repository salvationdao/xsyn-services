BEGIN;

CREATE TABLE fingerprints (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    visitor_id TEXT UNIQUE NOT NULL,
    os_cpu TEXT,
    platform TEXT,
    timezone TEXT,
    confidence DECIMAL,
    user_agent TEXT,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_fingerprints (
    user_id UUID NOT NULL REFERENCES users (id),
    fingerprint_id UUID NOT NULL REFERENCES fingerprints (id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, fingerprint_id)
);

CREATE TABLE fingerprint_ips (
    ip TEXT NOT NULL,
    fingerprint_id UUID NOT NULL REFERENCES fingerprints (id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ip, fingerprint_id)
);

COMMIT;
