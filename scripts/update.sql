ALTER TABLE users
ADD COLUMN "adm_ban" Boolean DEFAULT TRUE NOT NULL;

UPDATE users
SET adm_ban = (invited_by IS NULL);

CREATE OR REPLACE FUNCTION mindwell.allow_adm_upd() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD.invited_by <> NEW.invited_by) THEN
            NEW.adm_ban = false;
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER alw_adm_upd
    BEFORE UPDATE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.allow_adm_upd();

CREATE OR REPLACE FUNCTION mindwell.ban_adm() RETURNS VOID AS $$
    UPDATE users 
    SET adm_ban = true
    WHERE name IN (
        SELECT gs.name 
        FROM adm AS gs
        JOIN adm AS gf ON gf.grandfather = gs.name
        WHERE NOT gf.sent OR (gs.sent AND NOT gs.received)
    );
$$ LANGUAGE SQL;

DELETE FROM adm;
