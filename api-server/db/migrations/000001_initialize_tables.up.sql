-- This adaptation is released under the MIT License.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SEQUENCE epoch_seq INCREMENT BY 1 MAXVALUE 9 CYCLE;
CREATE OR REPLACE FUNCTION generate_object_id() RETURNS varchar AS $$
DECLARE
    time_component bigint;
    epoch_seq int;
    machine_id text := encode(gen_random_bytes(3), 'hex');
    process_id bigint;
    seq_id text := encode(gen_random_bytes(3), 'hex');
    result varchar:= '';
BEGIN
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp())) INTO time_component;
    SELECT nextval('epoch_seq') INTO epoch_seq;
    SELECT pg_backend_pid() INTO process_id;

    result := result || lpad(to_hex(time_component), 8, '0');
    result := result || machine_id;
    result := result || lpad(to_hex(process_id), 4, '0');
    result := result || seq_id;
    result := result || epoch_seq;
    RETURN result;
END;
$$ LANGUAGE PLPGSQL;

CREATE TYPE "user_perm" AS ENUM ('default', 'admin');
CREATE TYPE "member_role" AS ENUM ('guest', 'developer', 'admin');

CREATE TABLE IF NOT EXISTS "user" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    perm user_perm NOT NULL DEFAULT 'default',
    name VARCHAR(128) UNIQUE NOT NULL,
    first_name VARCHAR(128) NOT NULL,
    last_name VARCHAR(128) NOT NULL,
    email VARCHAR(256) UNIQUE DEFAULT NULL,
    password VARCHAR(1024) NOT NULL,
    config TEXT DEFAULT '{}',
    api_token VARCHAR(256) DEFAULT NULL,
    github_username VARCHAR(128) UNIQUE DEFAULT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS "organization" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) UNIQUE NOT NULL,
    description TEXT,
    config TEXT DEFAULT '{}',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS "user_group" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    organization_id INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_userGroup_orgId_name" ON "user_group" ("organization_id", "name");

CREATE TABLE IF NOT EXISTS "organization_member" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    organization_id INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    user_group_id INTEGER REFERENCES "user_group"("id") ON DELETE CASCADE,
    user_id INTEGER REFERENCES "user"("id") ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'guest',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_orgMember_orgId_userGroupId_userId" ON "organization_member" ("organization_id", "user_group_id", "user_id");

CREATE TABLE IF NOT EXISTS "user_group_user_relation" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    user_group_id INTEGER NOT NULL REFERENCES "user_group"("id") ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS "cluster" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    organization_id INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    kube_config TEXT NOT NULL,
    config TEXT DEFAULT '{}',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_cluster_orgId_name" ON "cluster" ("organization_id", "name");

CREATE TABLE IF NOT EXISTS "cluster_member" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    cluster_id INTEGER NOT NULL REFERENCES "cluster"("id") ON DELETE CASCADE,
    user_group_id INTEGER REFERENCES "user_group"("id") ON DELETE CASCADE,
    user_id INTEGER REFERENCES "user"("id") ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'guest',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_clusterMember_userGroupId_userId" ON "cluster_member" ("cluster_id", "user_group_id", "user_id");

CREATE TABLE IF NOT EXISTS "bento" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    manifest TEXT,
    organization_id INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_bento_orgId_name" ON bento ("organization_id", "name");

CREATE TYPE "bento_version_upload_status" AS ENUM ('pending', 'uploading', 'success', 'failed');
CREATE TYPE "bento_version_image_build_status" AS ENUM ('pending', 'building', 'success', 'failed');

CREATE TABLE IF NOT EXISTS "bento_version" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    version VARCHAR(512) NOT NULL,
    description TEXT,
    manifest TEXT,
    file_path TEXT,
    bento_id INTEGER NOT NULL REFERENCES bento("id") ON DELETE CASCADE,
    upload_status bento_version_upload_status NOT NULL DEFAULT 'pending',
    image_build_status bento_version_image_build_status NOT NULL DEFAULT 'pending',
    image_build_status_syncing_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    image_build_status_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    upload_started_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    upload_finished_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    upload_finished_reason TEXT,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    build_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_bentoVersion_bentoId_version" ON "bento_version" ("bento_id", "version");

CREATE TABLE IF NOT EXISTS "model" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    manifest TEXT,
    organization_id INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_model_orgId_name" ON "model" ("organization_id", "name");

CREATE TYPE "model_version_upload_status" AS ENUM ('pending', 'uploading', 'success', 'failed');
CREATE TYPE "model_version_image_build_status" AS ENUM ('pending', 'building', 'success', 'failed');

CREATE TABLE IF NOT EXISTS "model_version" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    version VARCHAR(512) NOT NULL,
    description TEXT,
    model_id INTEGER NOT NULL REFERENCES "model"("id") ON DELETE CASCADE,
    upload_status model_version_upload_status NOT NULL DEFAULT 'pending',
    image_build_status model_version_image_build_status NOT NULL DEFAULT 'pending',
    image_build_status_syncing_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    image_build_status_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    upload_started_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    upload_finished_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    upload_finished_reason TEXT,
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    build_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    manifest TEXT
);

CREATE UNIQUE INDEX "uk_modelVersion_modelId_version" ON "model_version" ("model_id", "version");

CREATE TYPE "deployment_status" AS ENUM ('unknown', 'non-deployed', 'failed', 'unhealthy', 'deploying', 'running');

CREATE TABLE IF NOT EXISTS "deployment" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    description TEXT,
    cluster_id INTEGER NOT NULL REFERENCES "cluster"("id") ON DELETE CASCADE,
    status deployment_status NOT NULL DEFAULT 'non-deployed',
    status_syncing_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    status_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    "kube_deploy_token" VARCHAR(128) DEFAULT '',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_deployment_clusterId_name" ON "deployment" ("cluster_id", "name");

CREATE TYPE "deployment_snapshot_type" AS ENUM ('stable', 'canary');
CREATE TYPE "deployment_snapshot_status" AS ENUM ('active', 'inactive');

CREATE TABLE IF NOT EXISTS "deployment_snapshot" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    type deployment_snapshot_type NOT NULL DEFAULT 'stable',
    status deployment_snapshot_status NOT NULL DEFAULT 'active',
    canary_rules TEXT,
    deployment_id INTEGER NOT NULL REFERENCES "deployment"("id") ON DELETE CASCADE,
    bento_version_id INTEGER REFERENCES "bento_version"("id") ON DELETE CASCADE,
    config TEXT DEFAULT '{}',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TYPE "resource_type" AS ENUM ('organization', 'cluster', 'bento', 'bento_version', 'deployment', 'deployment_snapshot', 'model', 'model_version');

CREATE TABLE IF NOT EXISTS "event" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    name VARCHAR(128) NOT NULL,
    organization_id INTEGER REFERENCES "organization"("id") ON DELETE CASCADE,
    cluster_id INTEGER REFERENCES "cluster"("id") ON DELETE CASCADE,
    resource_type resource_type NOT NULL,
    resource_id INTEGER NOT NULL,
    operation_name VARCHAR(128) NOT NULL,
    info TEXT DEFAULT '{}',
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS "terminal_record" (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(32) UNIQUE NOT NULL DEFAULT generate_object_id(),
    organization_id INTEGER DEFAULT NULL REFERENCES "organization"("id") ON DELETE CASCADE,
    cluster_id INTEGER DEFAULT NULL REFERENCES "cluster"("id") ON DELETE CASCADE,
    deployment_id INTEGER DEFAULT NULL REFERENCES "deployment"("id") ON DELETE CASCADE,
    resource_type resource_type NOT NULL,
    resource_id INTEGER NOT NULL,
    pod_name VARCHAR(128) NOT NULL,
    container_name VARCHAR(128) NOT NULL,
    meta TEXT,
    content TEXT[],
    creator_id INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS "cache" (
    id SERIAL PRIMARY KEY,
    key VARCHAR(512) UNIQUE NOT NULL,
    value TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TYPE "label_resource_type" as ENUM ('bento', 'bento_version', 'deployment', 'deployment_snapshot', "model", "model_version");

CREATE TABLE IF NOT EXISTS "label" (
    id SERIAL PRIMARY KEY,
    resource_type label_resource_type NOT NULL,
    resource_id INTEGER NOT NULL,
    key VARCHAR(128) NOT NULL,
    value VARCHAR(128) NOT NULL,
    creator_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX "uk_resoure_type_id_key" on "label" ("resource_type", "resource_id", "key");