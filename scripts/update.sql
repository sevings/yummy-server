CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION to_search_string(name TEXT, show_name TEXT, country TEXT, city TEXT)
   RETURNS TEXT AS $$
BEGIN
  RETURN name || ' ' || show_name || ' ' || country || ' ' || city;
END
$$ LANGUAGE 'plpgsql' IMMUTABLE;

CREATE INDEX "index_user_search" ON "mindwell"."users" USING GIST 
    (to_search_string("name", "show_name", "country", "city") gist_trgm_ops);
