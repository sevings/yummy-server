ALTER TABLE adm
ADD COLUMN "grandfather" Text NOT NULL DEFAULT '';

ALTER TABLE adm
ADD COLUMN "sent" Boolean NOT NULL DEFAULT false;

ALTER TABLE adm
ADD COLUMN "received" Boolean NOT NULL DEFAULT false;

CREATE OR REPLACE FUNCTION give_invites() RETURNS VOID AS $$
    WITH inviters AS (
        UPDATE mindwell.users 
        SET last_invite = CURRENT_DATE
        WHERE (rank <= (
                    SELECT COUNT(DISTINCT author_id) / 2
                    FROM mindwell.entries
                    WHERE age(created_at) <= '2 months'
                        AND visible_for = (SELECT id FROM entry_privacy WHERE type = 'all')
                )
                AND age(last_invite) >= interval '7 days'
                AND (SELECT COUNT(*) FROM mindwell.invites WHERE referrer_id = users.id) < 3
            ) OR (
                last_invite = created_at::Date
                AND (
                    SELECT COUNT(DISTINCT entries.id)
                    FROM mindwell.entries, mindwell.entry_votes
                    WHERE entries.author_id = users.id AND entry_votes.entry_id = entries.id
                        AND entry_votes.vote > 0 AND entry_votes.user_id <> users.invited_by
                ) >= 10
            )
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
