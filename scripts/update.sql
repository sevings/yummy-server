DROP TRIGGER cnt_invited ON users;
DROP FUNCTION count_invited();

CREATE OR REPLACE FUNCTION mindwell.count_invited_upd() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD.invited_by = NEW.invited_by) THEN
            RETURN NULL;
        END IF;

        UPDATE mindwell.users
        SET invited_count = invited_count + 1
        WHERE id = NEW.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_upd
    AFTER UPDATE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_invited_upd();

CREATE OR REPLACE FUNCTION mindwell.count_invited_del() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD.invited_by IS NULL) THEN
            RETURN NULL;
        END IF;

        UPDATE mindwell.users
        SET invited_count = invited_count - 1
        WHERE id = OLD.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_del
    AFTER DELETE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_invited_del();

UPDATE users
SET invited_count = (
    SELECT count(*) 
    FROM users AS invited 
    WHERE invited.invited_by = users.id
);
