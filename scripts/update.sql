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
        SELECT ek.id, (users.karma * 4 + ek.karma + ck.karma) / 5 AS karma
        FROM mindwell.users, (
            SELECT users.id, sum(entry_votes.vote) AS karma
            FROM mindwell.entry_votes, mindwell.entries, mindwell.users
            WHERE abs(entry_votes.vote) > 0.2 AND age(entries.created_at) <= interval '2 months'
                AND entry_votes.entry_id = entries.id AND entries.author_id = users.id
            GROUP BY users.id  
        ) as ek, (
            SELECT users.id, sum(comment_votes.vote) / 10 AS karma
            FROM mindwell.comment_votes, mindwell.comments, mindwell.users
            WHERE abs(comment_votes.vote) > 0.2 AND age(comments.created_at) <= interval '2 months'
                AND comment_votes.comment_id = comments.id AND comments.author_id = users.id
            GROUP BY users.id  
        ) AS ck
        WHERE ek.id = ck.id AND ek.id = users.id
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

recalc_karma();
