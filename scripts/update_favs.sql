set search_path = mindwell;

ALTER TABLE favorites
ADD COLUMN "date" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL;

ALTER TABLE watching
ADD COLUMN "date" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL;
