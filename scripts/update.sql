DELETE FROM notifications
WHERE CASE (SELECT "type" FROM notification_type WHERE notification_type.id = notifications."type")
    WHEN 'comment' THEN
        NOT EXISTS(SELECT 1 FROM comments WHERE comments.id = notifications.subject_id)
    WHEN 'invite' THEN
        FALSE
    ELSE 
        NOT EXISTS(SELECT 1 FROM users WHERE users.id = notifications.subject_id)
    END;

DELETE FROM notifications WHERE NOT EXISTS(SELECT 1 FROM users WHERE id = user_id);

ALTER TABLE notifications
ADD CONSTRAINT "notification_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id");

DELETE FROM images WHERE NOT EXISTS(SELECT 1 FROM users WHERE id = user_id);

ALTER TABLE images
ADD CONSTRAINT "image_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id");

CREATE OR REPLACE FUNCTION mindwell.delete_user(user_name TEXT) RETURNS VOID AS $$
    DECLARE
        user_id INTEGER;
    BEGIN
        user_id = (SELECT id FROM users WHERE lower(name) = lower(user_name));

        DELETE FROM mindwell.relations WHERE to_id = user_id;
        DELETE FROM mindwell.relations WHERE from_id = user_id;

        DELETE FROM mindwell.favorites WHERE favorites.user_id = delete_user.user_id;
        DELETE FROM mindwell.watching WHERE watching.user_id = delete_user.user_id;
        DELETE FROM mindwell.entries_privacy WHERE entries_privacy.user_id = delete_user.user_id;
        
        DELETE FROM mindwell.entry_votes WHERE entry_votes.user_id = delete_user.user_id;
        DELETE FROM mindwell.comment_votes WHERE comment_votes.user_id = delete_user.user_id;
        DELETE FROM mindwell.vote_weights WHERE vote_weights.user_id = delete_user.user_id;

        DELETE FROM mindwell.notifications
        WHERE nofitications.user_id = delete_user.user_id OR
            CASE (SELECT "type" FROM notification_type WHERE notification_type.id = notifications."type")
            WHEN 'comment' THEN
                (SELECT author_id FROM comments WHERE comments.id = notifications.subject_id) = delete_user.user_id
            WHEN 'invite' THEN
                FALSE
            ELSE 
                notifications.subject_id = delete_user.user_id
            END;

        DELETE FROM mindwell.images WHERE images.user_id = delete_user.user_id;
        DELETE FROM mindwell.entries WHERE author_id = user_id;
        DELETE FROM mindwell.comments WHERE author_id = user_id;
        DELETE FROM mindwell.users WHERE id = user_id;
    END;
$$ LANGUAGE plpgsql;
