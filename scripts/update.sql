
-- CREATE TABLE "complain_type" -------------------------------
CREATE TABLE "mindwell"."complain_type" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."complain_type" VALUES(0, 'comment');
INSERT INTO "mindwell"."complain_type" VALUES(1, 'entry');
-- -------------------------------------------------------------

-- CREATE TABLE "complains" ------------------------------------
CREATE TABLE "mindwell"."complains" (
    "id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "type" Integer NOT NULL,
    "subject_id" Integer NOT NULL,
    "content" Text NOT NULL,
	CONSTRAINT "unique_complain_id" PRIMARY KEY("id"),
    CONSTRAINT "complain_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "enum_complain_type" FOREIGN KEY("type") REFERENCES "mindwell"."complain_type"("id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_complain_id" ------------------------
CREATE UNIQUE INDEX "index_complain_id" ON "mindwell"."complains" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE "index_complain_user_id" -------------------
CREATE INDEX "index_complain_user_id" ON "mindwell"."complains" USING btree( "user_id" );
-- -------------------------------------------------------------
