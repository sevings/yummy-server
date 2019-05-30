CREATE OR REPLACE FUNCTION mindwell.del_relation_from_ignored() RETURNS TRIGGER AS $$
    DECLARE
        ignored Integer;
    BEGIN
        ignored = (SELECT id FROM mindwell.relation WHERE "type" = 'ignored');

        IF (NEW."type" = ignored) THEN
            DELETE FROM relations
            WHERE relations.from_id = NEW.to_id 
                AND relations.to_id = NEW.from_id
                AND relations."type" != ignored;
        END IF;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER relation_from_ignored
    AFTER INSERT OR UPDATE ON mindwell.relations
    FOR EACH ROW EXECUTE PROCEDURE mindwell.del_relation_from_ignored();
