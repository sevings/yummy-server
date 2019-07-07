DROP VIEW feed;
DROP VIEW short_users;
DROP VIEW long_users;

CREATE OR REPLACE FUNCTION mindwell.is_online(last_seen_at Timestamp With Time Zone) RETURNS BOOLEAN AS $$
    BEGIN
        RETURN now() - last_seen_at < interval '15 minutes';
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.user_age(birthday Date) RETURNS Integer AS $$
    BEGIN
        RETURN extract(year from age(birthday))::integer;
    END;
$$ LANGUAGE plpgsql;
