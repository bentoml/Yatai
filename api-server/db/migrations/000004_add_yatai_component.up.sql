BEGIN;
ALTER TYPE "resource_type" ADD VALUE 'yatai_component';
COMMIT;

BEGIN;
CREATE TABLE IF NOT EXISTS "yatai_component" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    version VARCHAR(128),
    cluster_id INTEGER NOT NULL REFERENCES "cluster"("id") ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    kube_namespace VARCHAR(128) NOT NULL,
    manifest JSONB,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    latest_heartbeat_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    latest_installed_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);
COMMIT;

BEGIN;
CREATE UNIQUE INDEX "uk_yataiComponent_clusterId_name" ON "yatai_component" ("cluster_id", "name");
COMMIT;
