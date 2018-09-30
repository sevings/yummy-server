CREATE OR REPLACE FUNCTION next_user_rank() RETURNS INTEGER AS $$
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
ADD COLUMN rank Integer DEFAULT next_user_rank() NOT NULL;

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
        SELECT id, row_number() OVER (ORDER BY karma DESC, created_at ASC) as rank
        FROM mindwell.users
    )
    UPDATE mindwell.users
    SET rank = upd.rank
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
    SELECT id, row_number() OVER (ORDER BY karma DESC, created_at ASC) as rank
    FROM mindwell.users
)
UPDATE mindwell.users
SET rank = upd.rank
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

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS mindwell.burn_karma();

ALTER TABLE users DROP COLUMN karma_raw;

DROP VIEW feed;
DROP VIEW long_users;

CREATE VIEW mindwell.long_users AS
SELECT users.id,
    users.name,
    users.show_name,
    users.password_hash,
    gender.type AS gender,
    users.is_daylog,
    user_privacy.type AS privacy,
    users.title,
    users.last_seen_at,
    users.rank,
    users.created_at,
    users.invited_by,
    users.birthday,
    users.css,
    users.entries_count,
    users.followings_count,
    users.followers_count,
    users.comments_count,
    users.ignored_count,
    users.invited_count,
    users.favorites_count,
    users.tags_count,
    extract(day from age(users.created_at))::integer AS "days_count",
    users.country,
    users.city,
    users.email,
    users.verified,
    users.api_key,
    users.valid_thru,
    users.avatar,
    users.cover,
    font_family.type AS font_family,
    users.font_size,
    alignment.type AS text_alignment,
    users.text_color,
    users.background_color,
    now() - last_seen_at < interval '15 minutes' AS is_online,
    extract(year from age(birthday))::integer as "age",
    short_users.id AS invited_by_id,
    short_users.name AS invited_by_name,
    short_users.show_name AS invited_by_show_name,
    short_users.is_online AS invited_by_is_online,
    short_users.avatar AS invited_by_avatar
FROM mindwell.users, mindwell.short_users,
    mindwell.gender, mindwell.user_privacy, mindwell.font_family, mindwell.alignment
WHERE users.invited_by = short_users.id
    AND users.gender = gender.id
    AND users.privacy = user_privacy.id
    AND users.font_family = font_family.id
    AND users.text_alignment = alignment.id;

CREATE VIEW mindwell.feed AS
SELECT entries.id, entries.created_at, rating, up_votes, down_votes,
    entries.title, cut_title, content, cut_content, edit_content, 
    has_cut, word_count,
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

CREATE OR REPLACE FUNCTION give_invites() RETURNS VOID AS $$
    WITH inviters AS (
        UPDATE mindwell.users 
        SET last_invite = CURRENT_DATE
        WHERE (rank <= (
                    SELECT COUNT(*) / 2
                    FROM mindwell.users
                    WHERE karma > 0
                )
                AND age(last_invite) >= interval '7 days'
                AND (SELECT COUNT(*) FROM mindwell.invites WHERE referrer_id = users.id) < 3
            ) OR (
                last_invite = created_at::Date
                AND (
                    SELECT COUNT(entries.id)
                    FROM mindwell.entries, mindwell.entry_votes
                    WHERE entries.author_id = users.id AND entry_votes.entry_id = entries.id
                        AND entry_votes.vote > 0 AND entry_votes.user_id <> users.invited_by
                ) >= 10
            )
        RETURNING users.id
    ), wc AS (
        SELECT COUNT(*) AS words FROM invite_words
    )
    INSERT INTO mindwell.invites(referrer_id, word1, word2, word3)
        SELECT inviters.id, 
            trunc(random() * wc.words),
            trunc(random() * wc.words),
            trunc(random() * wc.words)
        FROM inviters, wc
        ON CONFLICT (word1, word2, word3) DO NOTHING;
$$ LANGUAGE SQL;

ALTER TABLE comments
ADD COLUMN "edit_content" Text DEFAULT '' NOT NULL;

UPDATE comments
SET edit_content = content, content = '<p>' || content || '</p>';
