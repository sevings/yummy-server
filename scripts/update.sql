CREATE TABLE "mindwell"."size" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."size" VALUES(0, 'small');
INSERT INTO "mindwell"."size" VALUES(1, 'medium');
INSERT INTO "mindwell"."size" VALUES(2, 'large');
INSERT INTO "mindwell"."size" VALUES(3, 'thumbnail');

CREATE TABLE "mindwell"."image_sizes" (
    "image_id" Integer NOT NULL,
    "size" Integer NOT NULL,
    "width" Integer NOT NULL,
    "height" Integer NOT NULL,
    CONSTRAINT "unique_image_size" PRIMARY KEY ("image_id", "size"),
    CONSTRAINT "unique_image_id" FOREIGN KEY ("image_id") REFERENCES "mindwell"."images"("id") ON DELETE CASCADE,
    CONSTRAINT "enum_image_size" FOREIGN KEY("size") REFERENCES "mindwell"."size"("id")
);

CREATE INDEX "index_image_size_id" ON "mindwell"."image_sizes" USING btree( "image_id" );

CREATE TABLE "mindwell"."entry_images" (
    "entry_id" Integer NOT NULL,
    "image_id" Integer NOT NULL,
    CONSTRAINT "entry_images_entry" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "entry_images_image" FOREIGN KEY("image_id") REFERENCES "mindwell"."images"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_entry_image" UNIQUE("entry_id", "image_id") );
;

CREATE INDEX "index_entry_images_entry" ON "mindwell"."entry_images" USING btree( "entry_id" );
CREATE INDEX "index_entry_images_image" ON "mindwell"."entry_images" USING btree( "image_id" );

ALTER TABLE "mindwell"."images"
ADD COLUMN "extension" Text NOT NULL DEFAULT 'jpg';

ALTER TABLE "mindwell"."images"
ALTER COLUMN "extension" DROP DEFAULT;
