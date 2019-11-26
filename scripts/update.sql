DROP TRIGGER alw_adm_upd ON mindwell.users;

CREATE OR REPLACE FUNCTION mindwell.allow_adm_upd() RETURNS TRIGGER AS $$
    BEGIN
        NEW.adm_ban = false;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER alw_adm_upd
    BEFORE UPDATE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.invited_by IS NULL AND NEW.invited_by IS NOT NULL)
    EXECUTE PROCEDURE mindwell.allow_adm_upd();

DROP TRIGGER cnt_invited_upd ON mindwell.users;

CREATE OR REPLACE FUNCTION mindwell.count_invited_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET invited_count = invited_count + 1
        WHERE id = NEW.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_upd
    AFTER UPDATE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.invited_by IS NULL AND NEW.invited_by IS NOT NULL)
    EXECUTE PROCEDURE mindwell.count_invited_upd();

DROP TRIGGER cnt_invited_del ON mindwell.users;

CREATE OR REPLACE FUNCTION mindwell.count_invited_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET invited_count = invited_count - 1
        WHERE id = OLD.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_del
    AFTER DELETE ON mindwell.users
    FOR EACH ROW
    WHEN (OLD.invited_by IS NOT NULL)
    EXECUTE PROCEDURE mindwell.count_invited_del();

CREATE OR REPLACE FUNCTION mindwell.ban_adm() RETURNS VOID AS $$
    UPDATE users 
    SET adm_ban = true
    WHERE name IN (
        SELECT gs.name 
        FROM adm AS gs
        JOIN adm AS gf ON gf.grandfather = gs.name
        WHERE (NOT gf.sent AND NOT gf.received) OR (gs.sent AND NOT gs.received)
    );
$$ LANGUAGE SQL;
