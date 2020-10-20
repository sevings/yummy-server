CREATE EXTENSION IF NOT EXISTS rum;

CREATE OR REPLACE FUNCTION to_search_vector(title TEXT, content TEXT)
   RETURNS tsvector AS $$
BEGIN
  RETURN to_tsvector(title || '\n' || content);
END
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE INDEX "index_entry_search" ON "mindwell"."entries" USING rum
    (to_search_vector("title", "edit_content") rum_tsvector_ops);
