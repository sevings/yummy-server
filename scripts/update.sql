DROP INDEX index_comment_date;

ALTER TABLE entries
ADD COLUMN "last_comment" Integer;

ALTER TABLE entries
ADD CONSTRAINT "entry_last_comment_id" FOREIGN KEY("last_comment") REFERENCES "mindwell"."comments"("id");

CREATE OR REPLACE FUNCTION mindwell.set_last_comment_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET last_comment = NEW.id 
        WHERE id = NEW.entry_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER last_comments_ins
    AFTER INSERT ON mindwell.comments
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_last_comment_ins();

CREATE OR REPLACE FUNCTION mindwell.set_last_comment_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET last_comment = (
            SELECT max(comments.id)
            FROM comments
            WHERE entry_id = OLD.entry_id AND id <> OLD.id
        )
        WHERE last_comment = OLD.id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER last_comments_del
    BEFORE DELETE ON mindwell.comments
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_last_comment_del();

UPDATE entries
SET last_comment = (
    SELECT max(comments.id)
    FROM comments
    WHERE entry_id = entries.id
)
WHERE comments_count > 0;

CREATE INDEX "index_last_comment_id" ON "mindwell"."entries" USING btree( "last_comment" );
