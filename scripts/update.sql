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
