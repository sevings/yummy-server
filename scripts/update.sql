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
        SELECT users.id, (users.karma * 4 + COALESCE(ek.karma, 0) + COALESCE(ck.karma, 0) / 10) / 5 AS karma
        FROM mindwell.users
        LEFT JOIN (
            SELECT users.id, sum(for_votes.vote) + sum(by_votes.vote) / 5 AS karma
            FROM mindwell.users
            LEFT JOIN mindwell.entries ON entries.author_id = users.id AS by_entries
            LEFT JOIN mindwell.entry_votes ON entry_votes.entry_id = by_entries.id AS for_votes
            LEFT JOIN mindwell.entry_votes ON entry_votes.user_id = users.id AS by_votes
            WHERE abs(for_votes.vote) > 0.2 AND age(by_entries.created_at) <= interval '2 months' 
                AND age(by_votes.created_at) <= interval '2 months' 
            GROUP BY users.id  
        ) as ek ON users.id = ek.id
        LEFT JOIN (
            SELECT users.id, sum(for_votes.vote) + sum(by_votes.vote) / 5 AS karma
            FROM mindwell.users
            LEFT JOIN mindwell.comments ON comments.author_id = users.id AS by_comments
            LEFT JOIN mindwell.comment_votes ON comment_votes.comment_id = by_comments.id AS for_votes
            LEFT JOIN mindwell.comment_votes ON comment_votes.user_id = users.id AS by_votes
            WHERE abs(for_votes.vote) > 0.2 AND age(by_comments.created_at) <= interval '2 months' 
                AND age(by_votes.created_at) <= interval '2 months' 
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
