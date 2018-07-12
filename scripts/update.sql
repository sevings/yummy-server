CREATE OR REPLACE FUNCTION mindwell.inc_comments() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET comments_count = comments_count + 1
        WHERE id = NEW.author_id;
        
        UPDATE mindwell.entries
        SET comments_count = comments_count + 1
        WHERE id = NEW.entry_id;
        
        INSERT INTO mindwell.watching
        VALUES(NEW.author_id, NEW.entry_id)
        ON CONFLICT ON CONSTRAINT unique_user_watching DO NOTHING;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;
