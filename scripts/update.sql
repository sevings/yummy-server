CREATE OR REPLACE FUNCTION mindwell.delete_user(user_name TEXT) RETURNS VOID AS $$
    DECLARE
        del_id INTEGER;
    BEGIN
        del_id = (SELECT id FROM users WHERE lower(name) = lower(user_name));

        DELETE FROM mindwell.relations WHERE to_id = del_id;
        DELETE FROM mindwell.relations WHERE from_id = del_id;

        DELETE FROM mindwell.favorites WHERE favorites.user_id = del_id;
        DELETE FROM mindwell.watching WHERE watching.user_id = del_id;
        DELETE FROM mindwell.entries_privacy WHERE entries_privacy.user_id = del_id;

        DELETE FROM mindwell.entry_votes WHERE entry_votes.user_id = del_id;
        DELETE FROM mindwell.comment_votes WHERE comment_votes.user_id = del_id;
        DELETE FROM mindwell.vote_weights WHERE vote_weights.user_id = del_id;

        DELETE FROM mindwell.notifications
        WHERE notifications.user_id = del_id OR
            CASE (SELECT "type" FROM notification_type WHERE notification_type.id = notifications."type")
            WHEN 'comment' THEN
                (SELECT author_id FROM comments WHERE comments.id = notifications.subject_id) = del_id
            WHEN 'invite' THEN
                FALSE
            ELSE
                notifications.subject_id = del_id
            END;

        DELETE FROM complains WHERE user_id = del_id;

        DELETE FROM mindwell.images WHERE images.user_id = del_id;
        DELETE FROM mindwell.entries WHERE author_id = del_id;
        DELETE FROM mindwell.comments WHERE author_id = del_id;
        DELETE FROM mindwell.users WHERE id = del_id;
    END;
$$ LANGUAGE plpgsql;

CREATE  OR REPLACE FUNCTION mindwell.is_online(last_seen_at TIMESTAMP WITH TIME ZONE) RETURNS BOOLEAN AS $$
    BEGIN
        RETURN now() - last_seen_at < interval '5 minutes';
    END;
$$ LANGUAGE plpgsql;

INSERT INTO "mindwell"."user_privacy" VALUES(3, 'registered');

UPDATE users
SET privacy = (SELECT id FROM user_privacy WHERE type = 'registered')
WHERE privacy = (SELECT id FROM user_privacy WHERE type = 'all');

CREATE OR REPLACE FUNCTION mindwell.can_view_tlog(tlog_id INTEGER) RETURNS BOOLEAN AS $$
    BEGIN
        RETURN (
            SELECT user_privacy.type = 'all'
            FROM users
            LEFT JOIN user_privacy ON users.privacy = user_privacy.id
            WHERE users.id = tlog_id
        );
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.can_view_tlog(user_id INTEGER, tlog_id INTEGER) RETURNS BOOLEAN AS $$
    DECLARE
        privacy TEXT;
        is_ignored BOOLEAN;
        is_invited BOOLEAN;
        is_follower BOOLEAN;
    BEGIN
        IF user_id <= 0 THEN
            RETURN (SELECT can_view_tlog(tlog_id));
        END IF;

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
        WHEN 'registered' THEN
            RETURN TRUE;
        WHEN 'invited' THEN
            SELECT invited_by IS NOT NULL
            INTO is_invited
            FROM users
            WHERE users.id = user_id;

            RETURN is_invited;
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
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.can_view_entry(user_id INTEGER, entry_id INTEGER, author_id INTEGER, entry_privacy TEXT) RETURNS BOOLEAN AS $$
    DECLARE
        allowed BOOLEAN;
    BEGIN
        IF author_id = user_id THEN
            RETURN TRUE;
        END IF;

        IF entry_privacy = 'anonymous' THEN
            RETURN user_id > 0;
        END IF;

        allowed = (SELECT can_view_tlog(user_id, author_id));

        IF NOT allowed THEN
            RETURN FALSE;
        END IF;

        CASE entry_privacy
        WHEN 'all' THEN
            RETURN TRUE;
        WHEN 'some' THEN
            IF user_id > 0 THEN
                SELECT TRUE
                INTO allowed
                FROM entries_privacy
                WHERE entries_privacy.user_id = can_view_entry.user_id
                    AND entries_privacy.entry_id = can_view_entry.entry_id;

                allowed = COALESCE(allowed, FALSE);
                RETURN allowed;
            ELSE
                RETURN FALSE;
            END IF;
        WHEN 'me' THEN
            RETURN FALSE;
        ELSE
            RETURN FALSE;
        END CASE;
    END;
$$ LANGUAGE plpgsql;
