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
        
        RETURN OLD;
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

ALTER TABLE users
ADD COLUMN "invite_ban" Date DEFAULT CURRENT_DATE;

ALTER TABLE users
ADD COLUMN "vote_ban" Date DEFAULT CURRENT_DATE + interval '1 month';

UPDATE users
SET vote_ban = created_at + interval '1 month';

CREATE OR REPLACE FUNCTION mindwell.ban_invite(userName Text) RETURNS VOID AS $$
    DELETE FROM mindwell.invites
    WHERE referrer_id = (SELECT id FROM users WHERE lower(name) = lower(userName));

    UPDATE mindwell.users
    SET invite_ban = CURRENT_DATE + interval '1 month'
    WHERE lower(name) = lower(userName);
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION give_invites() RETURNS VOID AS $$
    WITH inviters AS (
        UPDATE mindwell.users 
        SET last_invite = CURRENT_DATE
        WHERE ((id IN (
                    SELECT author_id 
                    FROM (
                        SELECT created_at, author_id, rating 
                        FROM entries 
                        WHERE age(created_at) <= interval '1 month' 
                            AND visible_for = (SELECT id FROM entry_privacy WHERE type = 'all')
                        ORDER BY rating DESC 
                        LIMIT 100) AS e
                    WHERE created_at::date = CURRENT_DATE - 3
                )
                AND (SELECT COUNT(*) FROM mindwell.invites WHERE referrer_id = users.id) < 3
            ) OR (
                last_invite = created_at::Date
                AND (
                    SELECT COUNT(DISTINCT entries.id)
                    FROM mindwell.entries, mindwell.entry_votes
                    WHERE entries.author_id = users.id AND entry_votes.entry_id = entries.id
                        AND entry_votes.vote > 0 AND entry_votes.user_id <> users.invited_by
                ) >= 10
            )) AND age(last_invite) >= interval '14 days'
            AND invite_ban <= CURRENT_DATE
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

CREATE OR REPLACE FUNCTION ban_user(userName TEXT) RETURNS TEXT AS $$
    BEGIN
        UPDATE mindwell.users
        SET api_key = (
            SELECT array_to_string(array(
                SELECT substr('abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789', 
                    trunc(random() * 62)::integer + 1, 1)
                FROM generate_series(1, 32)), '')
            ),
            password_hash = '', verified = false
        WHERE lower(users.name) = lower(userName);

        RETURN (SELECT name FROM mindwell.users WHERE id = (
            SELECT invited_by FROM users WHERE lower(name) = lower(userName)
        ));
    END;
$$ LANGUAGE plpgsql;

DROP view mindwell.feed;
DROP view mindwell.long_users;

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
    CURRENT_DATE - created_at::date AS "days_count",
    users.country,
    users.city,
    users.email,
    users.verified,
    users.api_key,
    users.valid_thru,
    users.invite_ban,
    users.vote_ban,
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

ALTER TABLE users
ADD COLUMN "telegram" Integer;

CREATE UNIQUE INDEX "index_telegram" ON "mindwell"."users" USING btree( "telegram" );
