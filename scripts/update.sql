CREATE OR REPLACE FUNCTION can_view_tlog(user_id INTEGER, tlog TEXT) RETURNS BOOLEAN AS $$
    DECLARE
        tlog_id INTEGER;
    BEGIN
        SELECT users.id
        INTO tlog_id
		FROM users
		WHERE lower(users.name) = lower(tlog);

        RETURN (SELECT can_view_tlog(user_id, tlog_id));
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION can_view_tlog(user_id INTEGER, tlog_id INTEGER) RETURNS BOOLEAN AS $$
    DECLARE
        privacy TEXT;
        is_ignored BOOLEAN;
        is_invited BOOLEAN;
        is_follower BOOLEAN;
    BEGIN
        IF user_id = tlog_id THEN
            RETURN TRUE;
        END IF;

        SELECT relation.type = 'ignored'
        INTO is_ignored
		FROM relations
		INNER JOIN relation ON relation.id = relations.type
		WHERE from_id = tlog_id AND to_id = user_id;

        IF is_ignored THEN
            RETURN FALSE;
        END IF;

        SELECT user_privacy.type
        INTO privacy
		FROM users
        LEFT JOIN user_privacy ON users.privacy = user_privacy.id
		WHERE users.id = tlog_id;

        CASE privacy
        WHEN 'all' THEN 
            RETURN TRUE;
        WHEN 'invited' THEN
            SELECT invited_by IS NOT NULL 
            INTO is_invited
            FROM users
            WHERE users.id = user_id;

            RETURn is_nvited;
        WHEN 'followers' THEN
            SELECT relation.type = 'followed'
            INTO is_follower
            FROM relations
            INNER JOIN relation ON relation.id = relations.type
            WHERE from_id = user_id AND to_id = tlog_id;

            is_follower = COALESCE(is_follower, FALSE);
            RETURN is_follower;
        ELSE
            RETURN FALSE;
        END CASE;

        RETURN FALSE;
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION can_view_entry(user_id INTEGER, entry_id INTEGER) RETURNS BOOLEAN AS $$
    DECLARE
        author_id INTEGER;
        entry_privacy TEXT;
        allowed BOOLEAN;
    BEGIN
		SELECT entries.author_id, entry_privacy.type
        INTO author_id, entry_privacy
		FROM entries
		INNER JOIN entry_privacy ON visible_for = entry_privacy.id
		WHERE entries.id = entry_id;

        IF author_id = user_id THEN 
            RETURN TRUE;
        END IF;

        IF entry_privacy = 'anonymous' THEN
            RETURN TRUE:
        END IF;

        allowed = (SELECT can_view_tlog(user_id, author_id));

        IF NOT allowed THEN
            RETURN FALSE;
        END IF;

        CASE entry_privacy
        WHEN 'all' THEN
            RETURN TRUE;
        WHEN 'some' THEN
            SELECT TRUE
            INTO allowed
			FROM entries_privacy 
			WHERE entries_privacy.user_id = user_id 
                AND entries_privacy.entry_id = entry_id;
            
            allowed = COALESCE(allowed, FALSE);
            RETURN allowed;
        WHEN 'me' THEN
            RETURN FALSE;
        ELSE
            RETURN FALSE;
        END CASE;

        RETURN FALSE;
    END;
$$ LANGUAGE plpgsql;
