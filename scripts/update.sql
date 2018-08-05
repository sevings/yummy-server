CREATE TABLE "mindwell"."images" (
	"id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
	"path" Text NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL);
 ;

CREATE UNIQUE INDEX "index_image_id" ON "mindwell"."images" USING btree( "id" );
