CREATE TABLE IF NOT EXISTS "label_configuration" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    organization_id INTEGER REFERENCES "organization"("id") ON DELETE CASCADE,
    key VARCHAR(128) NOT NULL,
    info TEXT NOT NULL,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_labelConfiguration_orgId_key" on "label_configuration" ("organization_id", "key");
