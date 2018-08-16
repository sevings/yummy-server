ALTER TABLE comments 
DROP COLUMN "rating";

ALTER TABLE comments 
ADD COLUMN "rating" Real DEFAULT 0 NOT NULL;

ALTER TABLE comments 
ADD COLUMN "up_votes" Integer DEFAULT 0 NOT NULL;

ALTER TABLE comments 
ADD COLUMN "down_votes" Integer DEFAULT 0 NOT NULL;

ALTER TABLE comments 
ADD COLUMN "vote_sum" Real DEFAULT 0 NOT NULL;

ALTER TABLE comments 
ADD COLUMN "weight_sum" Real DEFAULT 0 NOT NULL;

ALTER TABLE comment_votes
DROP COLUMN "positive";

ALTER TABLE comment_votes
DROP COLUMN "taken";

ALTER TABLE comment_votes
ADD COLUMN "vote" Real NOT NULL;

DROP FUNCTION inc_comment_votes CASCADE
DROP FUNCTION dec_comment_votes CASCADE
DROP FUNCTION inc_comment_votes2 CASCADE
DROP FUNCTION dec_comment_votes2 CASCADE

CREATE OR REPLACE FUNCTION mindwell.comment_votes_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET up_votes = up_votes + (NEW.vote > 0)::int,
            down_votes = down_votes + (NEW.vote < 0)::int,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            rating = atan2(weight_sum + abs(NEW.vote), 2)
                * (vote_sum + NEW.vote) / (weight_sum + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.comment_id;
        
        WITH cmnt AS (
            SELECT author_id
            FROM mindwell.comments
            WHERE id = NEW.comment_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count + 1,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            weight = atan2(vote_count + 1, 20) * (vote_sum + NEW.vote) 
                / (weight_sum + abs(NEW.vote)) / pi() * 2
        FROM cmnt
        WHERE user_id = cmnt.author_id 
            AND vote_weights.category = 
                (SELECT id FROM categories WHERE "type" = 'comment');

        IF abs(NEW.vote) > 0.2 THEN
            WITH cmnt AS (
                SELECT author_id
                FROM mindwell.comments
                WHERE id = NEW.comment_id
            )
            UPDATE mindwell.users
            SET karma = karma + NEW.vote
            FROM cmnt
            WHERE users.id = cmnt.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.comment_votes_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET up_votes = up_votes - (OLD.vote > 0)::int + (NEW.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int + (NEW.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            rating = atan2(weight_sum - abs(OLD.vote) + abs(NEW.vote), 2)
                * (vote_sum - OLD.vote + NEW.vote) / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.comment_id;
        
        WITH cmnt AS (
            SELECT author_id
            FROM mindwell.comments
            WHERE id = NEW.comment_id
        )
        UPDATE mindwell.vote_weights
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            weight = atan2(vote_count, 20) * (vote_sum - OLD.vote + NEW.vote) 
                / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 2
        FROM cmnt
        WHERE user_id = cmnt.author_id
            AND vote_weights.category = 
                (SELECT id FROM categories WHERE "type" = 'comment');

        IF abs(OLD.vote) > 0.2 THEN
            WITH cmnt AS (
                SELECT author_id
                FROM mindwell.comments
                WHERE id = OLD.comment_id
            )
            UPDATE mindwell.users
            SET karma = karma - OLD.vote
            FROM cmnt
            WHERE users.id = cmnt.author_id;
        END IF;

        IF abs(NEW.vote) > 0.2 THEN
            WITH cmnt AS (
                SELECT author_id
                FROM mindwell.comments
                WHERE id = NEW.comment_id
            )
            UPDATE mindwell.users
            SET karma = karma + NEW.vote
            FROM cmnt
            WHERE users.id = cmnt.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.comment_votes_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET up_votes = up_votes - (OLD.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            rating = CASE WHEN weight_sum = abs(OLD.vote) THEN 0
                ELSE atan2(weight_sum - abs(OLD.vote), 2)
                    * (vote_sum - OLD.vote) / (weight_sum - abs(OLD.vote)) / pi() * 200
                END
        WHERE id = OLD.comment_id;
        
        WITH cmnt AS (
            SELECT author_id
            FROM mindwell.comments
            WHERE id = OLD.comment_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count - 1,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            weight = CASE WHEN weight_sum = abs(OLD.vote) THEN 0.1
                ELSE atan2(vote_count - 1, 20) * (vote_sum - OLD.vote) 
                    / (weight_sum - abs(OLD.vote)) / pi() * 2
                END
        FROM cmnt
        WHERE user_id = cmnt.author_id
            AND vote_weights.category = 
                (SELECT id FROM categories WHERE "type" = 'comment');

        IF abs(OLD.vote) > 0.2 THEN
            WITH cmnt AS (
                SELECT author_id
                FROM mindwell.comments
                WHERE id = OLD.comment_id
            )
            UPDATE mindwell.users
            SET karma = karma - OLD.vote
            FROM cmnt
            WHERE users.id = cmnt.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_comment_votes_ins
    AFTER INSERT ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.comment_votes_ins();

CREATE TRIGGER cnt_comment_votes_upd
    AFTER UPDATE ON mindwell.comment_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.comment_votes_upd();

CREATE TRIGGER cnt_comment_votes_del
    AFTER DELETE ON mindwell.comment_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.comment_votes_del();
