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
