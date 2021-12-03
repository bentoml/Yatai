DROP TABLE IF EXISTS "label";
DROP TABLE IF EXISTS "cache";
DROP TABLE IF EXISTS "terminal_record";
DROP TABLE IF EXISTS "event";
DROP TYPE IF EXISTS "resource_type";
DROP TABLE IF EXISTS "deployment_target";
DROP TYPE IF EXISTS "deployment_target_type";
DROP TABLE IF EXISTS "deployment_revision";
DROP TYPE IF EXISTS "deployment_revision_status";
DROP TABLE IF EXISTS "deployment";
DROP TYPE IF EXISTS "deployment_status";
DROP TABLE IF EXISTS "bento_version_model_version_rel";
DROP TABLE IF EXISTS "bento_version";
DROP TYPE IF EXISTS "bento_version_image_build_status";
DROP TYPE IF EXISTS "bento_version_upload_status";
DROP TABLE IF EXISTS "bento";
DROP TABLE IF EXISTS "model_version";
DROP TYPE IF EXISTS "model_version_upload_status";
DROP TYPE IF EXISTS "model_version_image_build_status";
DROP TABLE IF EXISTS "model";
DROP TABLE IF EXISTS "cluster_member";
DROP TABLE IF EXISTS "cluster";
DROP TABLE IF EXISTS "user_group_user_relation";
DROP TABLE IF EXISTS "organization_member";
DROP TABLE IF EXISTS "user_group";
DROP TABLE IF EXISTS "api_token";
DROP TABLE IF EXISTS "organization";
DROP TABLE IF EXISTS "user";
DROP TYPE IF EXISTS "member_role";
DROP TYPE IF EXISTS "user_perm";
DROP SEQUENCE IF EXISTS "epoch_seq";
DROP FUNCTION IF EXISTS "generate_object_id()";
DROP TABLE IF EXISTS "schema_migrations";
