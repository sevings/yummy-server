DROP TRIGGER alw_adm_upd ON mindwell.users;

CREATE OR REPLACE FUNCTION mindwell.allow_adm_upd() RETURNS TRIGGER AS $$
    BEGIN
        NEW.adm_ban = false;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER alw_adm_upd
    BEFORE UPDATE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.invited_by IS NULL AND NEW.invited_by IS NOT NULL)
    EXECUTE PROCEDURE mindwell.allow_adm_upd();

DROP TRIGGER cnt_invited_upd ON mindwell.users;

CREATE OR REPLACE FUNCTION mindwell.count_invited_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET invited_count = invited_count + 1
        WHERE id = NEW.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_upd
    AFTER UPDATE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.invited_by IS NULL AND NEW.invited_by IS NOT NULL)
    EXECUTE PROCEDURE mindwell.count_invited_upd();

DROP TRIGGER cnt_invited_del ON mindwell.users;

CREATE OR REPLACE FUNCTION mindwell.count_invited_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET invited_count = invited_count - 1
        WHERE id = OLD.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_del
    AFTER DELETE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.invited_by IS NOT NULL)
    EXECUTE PROCEDURE mindwell.count_invited_del();

CREATE OR REPLACE FUNCTION mindwell.ban_adm() RETURNS VOID AS $$
    UPDATE users 
    SET adm_ban = true
    WHERE name IN (
        SELECT gs.name 
        FROM adm AS gs
        JOIN adm AS gf ON gf.grandfather = gs.name
        WHERE (NOT gf.sent AND NOT gf.received) OR (gs.sent AND NOT gs.received)
    );
$$ LANGUAGE SQL;

CREATE OR REPLACE FUNCTION mindwell.recalc_karma() RETURNS VOID AS $$
    WITH upd AS (
        SELECT users.id, (
                users.karma * 4
                + COALESCE(fek.karma, 0) + COALESCE(bek.karma, 0)
                + COALESCE(fck.karma, 0) / 10 + COALESCE(bck.karma, 0) / 10
            ) / 5 AS karma
        FROM mindwell.users
        LEFT JOIN (
            SELECT users.id, sum(entry_votes.vote) AS karma
            FROM mindwell.users
            JOIN mindwell.entries ON entries.author_id = users.id
            JOIN mindwell.entry_votes ON entry_votes.entry_id = entries.id
            WHERE abs(entry_votes.vote) > 0.2 AND age(entries.created_at) <= interval '2 months'
            GROUP BY users.id
        ) AS fek ON users.id = fek.id -- votes for users entries
        LEFT JOIN (
            SELECT users.id, sum(entry_votes.vote) / 5 AS karma
            FROM mindwell.users
            JOIN mindwell.entry_votes ON entry_votes.user_id = users.id
            WHERE entry_votes.vote < 0 AND age(entry_votes.created_at) <= interval '2 months'
            GROUP BY users.id
        ) AS bek ON users.id = bek.id -- entry votes by users
        LEFT JOIN (
            SELECT users.id, sum(comment_votes.vote) AS karma
            FROM mindwell.users
            JOIN mindwell.comments ON comments.author_id = users.id
            JOIN mindwell.comment_votes ON comment_votes.comment_id = comments.id
            WHERE abs(comment_votes.vote) > 0.2 AND age(comments.created_at) <= interval '2 months'
            GROUP BY users.id
        ) AS fck ON users.id = fck.id -- votes for users comments
        LEFT JOIN (
            SELECT users.id, sum(comment_votes.vote) / 5 AS karma
            FROM mindwell.users
            JOIN mindwell.comment_votes ON comment_votes.user_id = users.id
            WHERE comment_votes.vote < 0 AND age(comment_votes.created_at) <= interval '2 months'
            GROUP BY users.id  
        ) AS bck ON users.id = bck.id -- comment votes by users
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
