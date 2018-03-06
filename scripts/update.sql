set search_path = mindwell;

ALTER TABLE entry_tags 
DROP CONSTRAINT entry_tags_entry;
ALTER TABLE entry_tags 
ADD CONSTRAINT "entry_tags_entry" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE;

ALTER TABLE favorites 
DROP CONSTRAINT favorite_entry_id;
ALTER TABLE favorites 
ADD CONSTRAINT "favorite_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE;

ALTER TABLE watching 
DROP CONSTRAINT watching_entry_id;
ALTER TABLE watching 
ADD CONSTRAINT "watching_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE;

ALTER TABLE entry_votes 
DROP CONSTRAINT entry_vote_entry_id;
ALTER TABLE entry_votes 
ADD CONSTRAINT "entry_vote_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE;

ALTER TABLE entries_privacy 
DROP CONSTRAINT entries_privacy_entry_id;
ALTER TABLE entries_privacy 
ADD CONSTRAINT "entries_privacy_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE;

ALTER TABLE comments 
DROP CONSTRAINT comment_entry_id;
ALTER TABLE comments 
ADD CONSTRAINT "comment_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE;

ALTER TABLE comment_votes 
DROP CONSTRAINT comment_vote_comment_id;
ALTER TABLE comment_votes 
ADD CONSTRAINT "comment_vote_comment_id" FOREIGN KEY("comment_id") REFERENCES "mindwell"."comments"("id") ON DELETE CASCADE;

CREATE OR REPLACE FUNCTION mindwell.count_tags() RETURNS TRIGGER AS $$
    BEGIN
        WITH authors AS 
        (
            SELECT DISTINCT author_id as id
            FROM mindwell.entries, changes
            WHERE entries.id = changes.entry_id
        )
        UPDATE mindwell.users
        SET tags_count = counts.cnt 
        FROM authors,
        (
            SELECT author_id, COUNT(tag_id) as cnt
            FROM mindwell.entries, mindwell.entry_tags, authors
            WHERE authors.id = entries.author_id AND entries.id = entry_tags.entry_id
            GROUP BY author_id
        ) AS counts
        WHERE authors.id = users.id AND counts.author_id = users.id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;
