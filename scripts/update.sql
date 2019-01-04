
-- CREATE TABLE "notification_type" ----------------------------
CREATE TABLE "mindwell"."notification_type" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."notification_type" VALUES(0, 'comment');
INSERT INTO "mindwell"."notification_type" VALUES(1, 'follower');
INSERT INTO "mindwell"."notification_type" VALUES(2, 'request');
INSERT INTO "mindwell"."notification_type" VALUES(3, 'accept');
-- -------------------------------------------------------------



-- CREATE TABLE "notifications" --------------------------------
CREATE TABLE "mindwell"."notifications" (
    "id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "type" Integer NOT NULL,
    "subject_id" Integer NOT NULL,
    "read" Boolean DEFAULT FALSE NOT NULL,
	CONSTRAINT "unique_notification_id" PRIMARY KEY("id"),
    CONSTRAINT "enum_notification_type" FOREIGN KEY("type") REFERENCES "mindwell"."notification_type"("id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_notification_id" ------------------------
CREATE UNIQUE INDEX "index_notification_id" ON "mindwell"."notifications" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_notification_user_id" -------------------
CREATE INDEX "index_notification_user_id" ON "mindwell"."notifications" USING btree( "user_id" );
-- -------------------------------------------------------------

ALTER TABLE users
ALTER COLUMN email_comments SET DEFAULT FALSE;

ALTER TABLE users
ALTER COLUMN email_followers SET DEFAULT FALSE;
