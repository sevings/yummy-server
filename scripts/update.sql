DROP TRIGGER cnt_tags_ins ON entry_tags;
DROP TRIGGER cnt_tags_del ON entry_tags;
DROP FUNCTION count_tags();

CREATE OR REPLACE FUNCTION mindwell.count_tags_ins() RETURNS TRIGGER AS $$
    BEGIN
        WITH authors AS
        (
            SELECT DISTINCT author_id as id
            FROM changes
            INNER JOIN mindwell.entries ON changes.entry_id = entries.id
        )
        UPDATE mindwell.users
        SET tags_count = counts.cnt
        FROM
        (
            SELECT authors.id, COUNT(DISTINCT tag_id) as cnt
            FROM authors
            LEFT JOIN mindwell.entries ON authors.id = entries.author_id
            LEFT JOIN mindwell.entry_privacy ON entries.visible_for = entry_privacy.id
            LEFT JOIN mindwell.entry_tags ON entries.id = entry_tags.entry_id
            WHERE (entry_privacy.type = 'all' OR entry_privacy.type IS NULL)
            GROUP BY authors.id
        ) AS counts
        WHERE counts.id = users.id
            AND tags_count <> counts.cnt;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_tags_ins
    AFTER INSERT ON mindwell.entry_tags
    REFERENCING NEW TABLE as changes
    FOR EACH STATEMENT EXECUTE PROCEDURE mindwell.count_tags_ins();

CREATE OR REPLACE FUNCTION mindwell.count_tags_upd() RETURNS TRIGGER AS $$
    BEGIN
        WITH authors AS
        (
            SELECT OLD.author_id AS id
        )
        UPDATE mindwell.users
        SET tags_count = counts.cnt
        FROM
        (
            SELECT authors.id, COUNT(DISTINCT tag_id) as cnt
            FROM authors
            LEFT JOIN mindwell.entries ON authors.id = entries.author_id
            LEFT JOIN mindwell.entry_privacy ON entries.visible_for = entry_privacy.id
            LEFT JOIN mindwell.entry_tags ON entries.id = entry_tags.entry_id
            WHERE (entry_privacy.type = 'all' OR entry_privacy.type IS NULL)
            GROUP BY authors.id
        ) AS counts
        WHERE users.id = counts.id
            AND tags_count <> counts.cnt;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_tags_upd
    AFTER UPDATE ON mindwell.entries
    FOR EACH ROW
    WHEN ( OLD.visible_for <> NEW.visible_for )
    EXECUTE PROCEDURE mindwell.count_tags_upd();

CREATE OR REPLACE FUNCTION mindwell.count_tags_del() RETURNS TRIGGER AS $$
    BEGIN
        WITH authors AS
        (
            SELECT DISTINCT author_id as id
            FROM changes
        )
        UPDATE mindwell.users
        SET tags_count = counts.cnt
        FROM
        (
            SELECT authors.id, COUNT(DISTINCT tag_id) as cnt
            FROM authors
            LEFT JOIN mindwell.entries ON authors.id = entries.author_id
            LEFT JOIN mindwell.entry_privacy ON entries.visible_for = entry_privacy.id
            LEFT JOIN mindwell.entry_tags ON entries.id = entry_tags.entry_id
            WHERE (entry_privacy.type = 'all' OR entry_privacy.type IS NULL)
            GROUP BY authors.id
        ) AS counts
        WHERE users.id = counts.id
            AND tags_count <> counts.cnt;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_tags_del
    AFTER DELETE ON mindwell.entries
    REFERENCING OLD TABLE AS changes
    FOR EACH STATEMENT
    EXECUTE PROCEDURE mindwell.count_tags_del();

UPDATE mindwell.users
SET tags_count = counts.cnt
FROM
(
    SELECT users.id, COUNT(DISTINCT tag_id) as cnt
    FROM users
    LEFT JOIN mindwell.entries ON users.id = entries.author_id
    LEFT JOIN mindwell.entry_privacy ON entries.visible_for = entry_privacy.id
    LEFT JOIN mindwell.entry_tags ON entries.id = entry_tags.entry_id
    WHERE (entry_privacy.type = 'all' OR entry_privacy.type IS NULL)
    GROUP BY users.id
) AS counts
WHERE counts.id = users.id
    AND tags_count <> counts.cnt;
