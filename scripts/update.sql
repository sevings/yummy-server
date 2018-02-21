CREATE OR REPLACE FUNCTION mindwell.entry_votes_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET votes = votes + sign(NEW.vote), 
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            rating = (vote_sum + NEW.vote) / (weight_sum + abs(NEW.vote)) * 10
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
            weight = atan2(vote_count + 1, 5) * (vote_sum + NEW.vote) 
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
            SET karma = karma + NEW.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET votes = votes - sign(OLD.vote) + sign(NEW.vote), 
            vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            rating = (vote_sum - OLD.vote + NEW.vote) / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) * 10
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            weight = atan2(vote_count, 5) * (vote_sum - OLD.vote + NEW.vote) 
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
            SET karma = karma - OLD.vote * 5
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
            SET karma = karma + NEW.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET votes = votes - sign(OLD.vote), 
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            rating = CASE WHEN weight_sum = abs(OLD.vote) THEN 0
                ELSE (vote_sum - OLD.vote) / (weight_sum - abs(OLD.vote)) * 10
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
                ELSE atan2(vote_count - 1, 5) * (vote_sum - OLD.vote) 
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
            SET karma = karma - OLD.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;
