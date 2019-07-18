DELETE FROM notifications
WHERE CASE (SELECT "type" FROM notification_type WHERE notification_type.id = notifications."type")
    WHEN 'comment' THEN
        NOT EXISTS(SELECT 1 FROM comments WHERE comments.id = notifications.subject_id)
    WHEN 'invite' THEN
        FALSE
    ELSE 
        NOT EXISTS(SELECT 1 FROM users WHERE users.id = notifications.subject_id)
    END;
