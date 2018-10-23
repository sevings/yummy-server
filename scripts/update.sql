CREATE TABLE "mindwell"."adm" (
	"name" Text NOT NULL,
    "fullname" Text NOT NULL,
    "postcode" Text NOT NULL,
    "country" Text NOT NULL,
    "address" Text NOT NULL,
    "comment" Text NOT NULL,
    "anonymous" Boolean NOT NULL );
;

CREATE UNIQUE INDEX "index_adm" ON "mindwell"."adm" USING btree( lower("name") );
