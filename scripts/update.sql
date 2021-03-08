CREATE TABLE "mindwell"."apps" (
    "id" Integer UNIQUE NOT NULL,
    "secret_hash" Bytea NOT NULL,
    "redirect_uri" Text NOT NULL,
    "developer_id" Integer NOT NULL,
    "flow" Smallint NOT NULL,
    "name" Text NOT NULL,
    "show_name" Text NOT NULL,
    "platform" Text NOT NULL,
    "info" Text NOT NULL,
    "ban" Bool NOT NULL DEFAULT FALSE,
     CONSTRAINT "unique_app_id" PRIMARY KEY("id"),
     CONSTRAINT "app_developer_id" FOREIGN KEY("developer_id") REFERENCES "mindwell"."users"("id") ON DELETE CASCADE
);

CREATE TABLE "mindwell"."sessions" (
    "id" Bigserial,
    "app_id" Integer NOT NULL,
    "user_id" Integer NOT NULL,
    "scope" Integer NOT NULL,
    "access_hash" Bytea NOT NULL,
    "refresh_hash" Bytea NOT NULL,
    "access_thru" Timestamp With Time Zone NOT NULL,
    "refresh_thru" Timestamp With Time Zone NOT NULL,
    CONSTRAINT "session_id" PRIMARY KEY("id"),
    CONSTRAINT "session_app_id" FOREIGN KEY("app_id") REFERENCES "mindwell"."apps"("id") ON DELETE CASCADE,
    CONSTRAINT "session_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id") ON DELETE CASCADE
);

CREATE INDEX "index_access_hash" ON "mindwell"."sessions" USING btree( "access_hash" );
CREATE INDEX "index_refresh_hash" ON "mindwell"."sessions" USING btree( "refresh_hash" );

CREATE TABLE "mindwell"."app_tokens" (
    "app_id" Integer NOT NULL,
    "token_hash" Bytea NOT NULL,
    "valid_thru" Timestamp With Time Zone NOT NULL,
     CONSTRAINT "app_token_app_id" FOREIGN KEY("app_id") REFERENCES "mindwell"."apps"("id") ON DELETE CASCADE
);

CREATE INDEX "index_app_token_hash" ON "mindwell"."app_tokens" USING btree( "token_hash" );
