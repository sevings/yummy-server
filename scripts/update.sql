CREATE TABLE "mindwell"."images" (
	"id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
	"path" Text NOT NULL,
    "mime" Text NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL);
 ;

CREATE UNIQUE INDEX "index_image_id" ON "mindwell"."images" USING btree( "id" );

CREATE TABLE "mindwell"."size" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."size" VALUES(0, 'small');
INSERT INTO "mindwell"."size" VALUES(1, 'medium');
INSERT INTO "mindwell"."size" VALUES(2, 'large');

CREATE TABLE "mindwell"."image_sizes" (
    "image_id" Integer NOT NULL,
    "size" Integer NOT NULL,
    "width" Integer NOT NULL,
    "height" Integer NOT NULL,
    CONSTRAINT "unique_image_size" PRIMARY KEY ("image_id", "size"),
    CONSTRAINT "unique_image_id" FOREIGN KEY ("image_id") REFERENCES "mindwell"."images"("id"),
    CONSTRAINT "enum_image_size" FOREIGN KEY("size") REFERENCES "mindwell"."size"("id")
);

CREATE INDEX "index_image_size_id" ON "mindwell"."image_sizes" USING btree( "image_id" );
