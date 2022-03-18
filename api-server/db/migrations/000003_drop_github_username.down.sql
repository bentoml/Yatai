ALTER TABLE "user"
ADD COLUMN "github_username" varchar(128) UNIQUE DEFAULT NULL;