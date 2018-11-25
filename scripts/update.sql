ALTER TABLE adm
ADD COLUMN "grandfather" Text NOT NULL DEFAULT '';

ALTER TABLE adm
ADD COLUMN "sent" Boolean NOT NULL DEFAULT false;

ALTER TABLE adm
ADD COLUMN "received" Boolean NOT NULL DEFAULT false;
