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
