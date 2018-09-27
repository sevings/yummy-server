CREATE OR REPLACE FUNCTION next_user_position() RETURNS INTEGER AS $$
    DECLARE
        pos INTEGER;
    BEGIN
        pos = (
            SELECT COUNT(*) + 1
            FROM users
            WHERE karma >= 0
        );
        RETURN pos;
    END;
$$ language plpgsql;

ALTER TABLE users
ADD COLUMN position Integer DEFAULT next_user_position() NOT NULL;

CREATE OR REPLACE FUNCTION recalc_karma() RETURNS VOID AS $$
    WITH upd AS (
        SELECT users.id, (users.karma * 4 + COALESCE(ek.karma, 0) + COALESCE(ck.karma, 0)) / 5 AS karma
        FROM mindwell.users
        LEFT JOIN (
            SELECT users.id, sum(entry_votes.vote) AS karma
            FROM mindwell.entry_votes, mindwell.entries, mindwell.users
            WHERE abs(entry_votes.vote) > 0.2 AND age(entries.created_at) <= interval '2 months'
                AND entry_votes.entry_id = entries.id AND entries.author_id = users.id
            GROUP BY users.id  
        ) as ek ON users.id = ek.id
        LEFT JOIN (
            SELECT users.id, sum(comment_votes.vote) / 10 AS karma
            FROM mindwell.comment_votes, mindwell.comments, mindwell.users
            WHERE abs(comment_votes.vote) > 0.2 AND age(comments.created_at) <= interval '2 months'
                AND comment_votes.comment_id = comments.id AND comments.author_id = users.id
            GROUP BY users.id  
        ) AS ck ON users.id = ck.id
    )
    UPDATE mindwell.users
    SET karma = upd.karma
    FROM upd
    WHERE users.id = upd.id;

    WITH upd AS (
        SELECT id, row_number() OVER (ORDER BY karma DESC, created_at ASC) as position
        FROM mindwell.users
    )
    UPDATE mindwell.users
    SET position = upd.position
    FROM upd
    WHERE users.id = upd.id;
$$ LANGUAGE SQL;

WITH upd AS (
    SELECT users.id, COALESCE(ek.karma, 0) + COALESCE(ck.karma, 0) AS karma
    FROM mindwell.users
    LEFT JOIN (
        SELECT users.id, sum(entry_votes.vote) AS karma
        FROM mindwell.entry_votes, mindwell.entries, mindwell.users
        WHERE abs(entry_votes.vote) > 0.2 AND age(entries.created_at) <= interval '2 months'
            AND entry_votes.entry_id = entries.id AND entries.author_id = users.id
        GROUP BY users.id  
    ) as ek ON users.id = ek.id
    LEFT JOIN (
        SELECT users.id, sum(comment_votes.vote) / 10 AS karma
        FROM mindwell.comment_votes, mindwell.comments, mindwell.users
        WHERE abs(comment_votes.vote) > 0.2 AND age(comments.created_at) <= interval '2 months'
            AND comment_votes.comment_id = comments.id AND comments.author_id = users.id
        GROUP BY users.id  
    ) AS ck ON users.id = ck.id
)
UPDATE mindwell.users
SET karma = upd.karma
FROM upd
WHERE users.id = upd.id;

WITH upd AS (
    SELECT id, row_number() OVER (ORDER BY karma DESC, created_at ASC) as position
    FROM mindwell.users
)
UPDATE mindwell.users
SET position = upd.position
FROM upd
WHERE users.id = upd.id;

DROP FUNCTION IF EXISTS mindwell.upd_karma() CASCADE;

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
            SET karma = karma + NEW.vote
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
            SET karma = karma - OLD.vote
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
            SET karma = karma + NEW.vote
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
            SET karma = karma - OLD.vote
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
            SET karma = karma + NEW.vote / 10
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
            SET karma = karma - OLD.vote / 10
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
            SET karma = karma + NEW.vote / 10
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
            SET karma = karma - OLD.vote / 10
            FROM cmnt
            WHERE users.id = cmnt.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS mindwell.burn_karma();

ALTER TABLE users DROP COLUMN karma_raw;
