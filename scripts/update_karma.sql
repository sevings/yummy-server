set search_path = mindwell;

CREATE OR REPLACE FUNCTION mindwell.create_vote_weights() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 0);
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 1);
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 2);
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 3);

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER crt_vote_weights
    AFTER INSERT ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.create_vote_weights();



-- CREATE TABLE "categories" -----------------------------------
CREATE TABLE "mindwell"."categories" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."categories" VALUES(0, 'tweet');
INSERT INTO "mindwell"."categories" VALUES(1, 'longread');
INSERT INTO "mindwell"."categories" VALUES(2, 'media');
INSERT INTO "mindwell"."categories" VALUES(3, 'comment');
-- -------------------------------------------------------------



DROP VIEW feed;



ALTER TABLE entries
ALTER COLUMN rating Real DEFAULT 0 NOT NULL;

ALTER TABLE entries
ADD COLUMN votes Integer DEFAULT 0 NOT NULL;

ALTER TABLE entries
ADD COLUMN vote_sum Real DEFAULT 0 NOT NULL;

ALTER TABLE entries
ADD COLUMN weight_sum Real DEFAULT 0 NOT NULL;

ALTER TABLE entries
ADD COLUMN category Integer DEFAULT 1 NOT NULL;

ALTER TABLE entries
ALTER COLUMN category DROP DEFAULT;

ALTER TABLE entries
ADD CONSTRAINT "entry_category" FOREIGN KEY("category") REFERENCES "mindwell"."categories"("id");



CREATE OR REPLACE VIEW mindwell.feed AS
SELECT entries.id, entries.created_at, rating, votes,
    entries.title, content, edit_content, word_count,
    entry_privacy.type AS entry_privacy,
    is_votable, entries.comments_count,
    long_users.id AS author_id,
    long_users.name AS author_name, 
    long_users.show_name AS author_show_name,
    long_users.is_online AS author_is_online,
    long_users.avatar AS author_avatar,
    long_users.privacy AS author_privacy
FROM mindwell.long_users, mindwell.entries, mindwell.entry_privacy
WHERE long_users.id = entries.author_id 
    AND entry_privacy.id = entries.visible_for;



ALTER TABLE entry_votes
DROP COLUMN positive;

ALTER TABLE entry_votes
DROP COLUMN taken;

ALTER TABLE entry_votes
ADD COLUMN vote Real NOT NULL;



DROP FUNCTION inc_entry_votes_ins() CASCADE;
DROP FUNCTION dec_entry_votes_ins() CASCADE;
DROP FUNCTION inc_entry_votes2() CASCADE;
DROP FUNCTION dec_entry_votes2() CASCADE;
DROP FUNCTION inc_entry_votes_del() CASCADE;
DROP FUNCTION dec_entry_votes_del() CASCADE;



CREATE OR REPLACE FUNCTION mindwell.entry_votes_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET votes = votes + sign(NEW.vote), 
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            rating = vote_sum / weight_sum
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count + 1,
            vote_sum = vote_sum + NEW.vote + 1, -- always positive - (0, 2)
            weight_sum = weight_sum + abs(NEW.vote),
            weight = atan2(vote_count, 5) * vote_sum / weight_sum / pi() -- / 2 / (pi() / 2) => / pi()
        WHERE user_id = entry.author_id AND category = entry.category;

        UPDATE mindwell.users
        SET karma = karma + NEW.vote * 5
        WHERE id = NEW.user_id;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            rating = vote_sum / weight_sum
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            weight = atan2(vote_count, 5) * vote_sum / weight_sum / pi() -- / 2 / (pi() / 2) => / pi()
        WHERE user_id = entry.author_id AND category = entry.category;

        UPDATE mindwell.users
        SET karma = karma - OLD.vote * 5 + NEW.vote * 5
        WHERE id = NEW.user_id;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET votes = votes - sign(OLD.vote), 
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            rating = vote_sum / weight_sum
        WHERE id = OLD.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = OLD.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count - 1,
            vote_sum = vote_sum - OLD.vote - 1, -- always positive - (0, 2)
            weight_sum = weight_sum - abs(OLD.vote),
            weight = atan2(vote_count, 5) * vote_sum / weight_sum / pi() -- / 2 / (pi() / 2) => / pi()
        WHERE user_id = entry.author_id AND category = entry.category;

        UPDATE mindwell.users
        SET karma = karma - OLD.vote * 5
        WHERE id = OLD.user_id;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_entry_votes_ins
    AFTER INSERT ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.entry_votes_ins();

CREATE TRIGGER cnt_entry_votes_upd
    AFTER UPDATE ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.entry_votes_upd();

CREATE TRIGGER cnt_entry_votes_del
    AFTER DELETE ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.entry_votes_del();



-- CREATE TABLE "vote_weights" ---------------------------------
CREATE TABLE "mindwell"."vote_weights" (
	"user_id" Integer NOT NULL,
	"category" Integer NOT NULL,
    "weight" Real DEFAULT 0.1 NOT NULL,
    "vote_count" Integer DEFAULT 0 NOT NULL,
    "vote_sum" Real DEFAULT 0 NOT NULL,
    "weight_sum" Real DEFAULT 0 NOT NULL,
    CONSTRAINT "vote_weights_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "vote_weights_category" FOREIGN KEY("category") REFERENCES "mindwell"."categories"("id"),
    CONSTRAINT "unique_vote_weight" UNIQUE("user_id", "category") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_vote_weights" ---------------------------
CREATE INDEX "index_vote_weights" ON "mindwell"."vote_weights" USING btree( "user_id" );
-- -------------------------------------------------------------
