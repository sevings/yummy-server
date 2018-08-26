ALTER TABLE users
ADD COLUMN "karma_raw" Real DEFAULT 0 NOT NULL;

CREATE OR REPLACE FUNCTION mindwell.upd_karma() RETURNS TRIGGER AS $$
    BEGIN
        NEW.karma = 
            CASE
                WHEN abs(NEW.karma_raw) < 1000 THEN sin(NEW.karma_raw * pi() / 2 / 1000) * 100
                WHEN NEW.karma_raw >= 1000 THEN 100
                ELSE -100
            END;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER upd_user_karma
    BEFORE UPDATE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.karma_raw IS DISTINCT FROM NEW.karma_raw)
    EXECUTE PROCEDURE mindwell.upd_karma();

CREATE OR REPLACE FUNCTION burn_karma() RETURNS VOID AS $$
    UPDATE mindwell.users
    SET karma_raw = 
        CASE 
            WHEN abs(karma_raw) > 25 THEN karma_raw * 0.98
            WHEN abs(karma_raw) > 0.5 THEN karma_raw - karma_raw / trunc(karma_raw * 2)
            ELSE 0
        END
    WHERE karma_raw <> 0;
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET up_votes = up_votes + (NEW.vote > 0)::int,
            down_votes = down_votes + (NEW.vote < 0)::int,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            rating = atan2(weight_sum + abs(NEW.vote), 2)
                * (vote_sum + NEW.vote) / (weight_sum + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count + 1,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            weight = atan2(vote_count + 1, 20) * (vote_sum + NEW.vote) 
                / (weight_sum + abs(NEW.vote)) / pi() * 2
        FROM entry
        WHERE user_id = entry.author_id 
            AND vote_weights.category = entry.category;

        IF abs(NEW.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = NEW.entry_id
            )
            UPDATE mindwell.users
            SET karma_raw = karma_raw + NEW.vote / 2
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET up_votes = up_votes - (OLD.vote > 0)::int + (NEW.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int + (NEW.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            rating = atan2(weight_sum - abs(OLD.vote) + abs(NEW.vote), 2)
                * (vote_sum - OLD.vote + NEW.vote) / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            weight = atan2(vote_count, 20) * (vote_sum - OLD.vote + NEW.vote) 
                / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 2
        FROM entry
        WHERE user_id = entry.author_id
            AND vote_weights.category = entry.category;

        IF abs(OLD.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = OLD.entry_id
            )
            UPDATE mindwell.users
            SET karma_raw = karma_raw - OLD.vote / 2
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        IF abs(NEW.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = NEW.entry_id
            )
            UPDATE mindwell.users
            SET karma_raw = karma_raw + NEW.vote / 2
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET up_votes = up_votes - (OLD.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            rating = CASE WHEN weight_sum = abs(OLD.vote) THEN 0
                ELSE atan2(weight_sum - abs(OLD.vote), 2)
                    * (vote_sum - OLD.vote) / (weight_sum - abs(OLD.vote)) / pi() * 200
                END
        WHERE id = OLD.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = OLD.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count - 1,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            weight = CASE WHEN weight_sum = abs(OLD.vote) THEN 0.1
                ELSE atan2(vote_count - 1, 20) * (vote_sum - OLD.vote) 
                    / (weight_sum - abs(OLD.vote)) / pi() * 2
                END
        FROM entry
        WHERE user_id = entry.author_id
            AND vote_weights.category = entry.category;

        IF abs(OLD.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = OLD.entry_id
            )
            UPDATE mindwell.users
            SET karma_raw = karma_raw - OLD.vote / 2
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

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
            SET karma_raw = karma_raw + NEW.vote / 20
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
            SET karma_raw = karma_raw - OLD.vote / 20
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
            SET karma_raw = karma_raw + NEW.vote / 20
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
            SET karma_raw = karma_raw - OLD.vote / 20
            FROM cmnt
            WHERE users.id = cmnt.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;
