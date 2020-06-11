ALTER TABLE entries
DROP COLUMN cut_title;

ALTER TABLE entries
DROP COLUMN content;

ALTER TABLE entries
DROP COLUMN cut_content;

ALTER TABLE entries
DROP COLUMN has_cut;

ALTER TABLE comments
DROP COLUMN content;
