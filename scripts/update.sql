ALTER TABLE users 
ADD COLUMN "comment_ban" Date DEFAULT CURRENT_DATE NOT NULL;

ALTER TABLE users 
ADD COLUMN "live_ban" Date DEFAULT CURRENT_DATE NOT NULL;

DROP VIEW feed;
DROP VIEW long_users;

CREATE VIEW mindwell.long_users AS
SELECT users.id,
    users.name,
    users.show_name,
    users.password_hash,
    gender.type AS gender,
    users.is_daylog,
    user_privacy.type AS privacy,
    users.title,
    users.last_seen_at,
    users.rank,
    users.created_at,
    users.invited_by,
    users.birthday,
    users.css,
    users.entries_count,
    users.followings_count,
    users.followers_count,
    users.comments_count,
    users.ignored_count,
    users.invited_count,
    users.favorites_count,
    users.tags_count,
    CURRENT_DATE - created_at::date AS "days_count",
    users.country,
    users.city,
    users.email,
    users.verified,
    users.api_key,
    users.valid_thru,
    users.invite_ban,
    users.vote_ban,
    users.comment_ban,
    users.live_ban,
    users.avatar,
    users.cover,
    font_family.type AS font_family,
    users.font_size,
    alignment.type AS text_alignment,
    users.text_color,
    users.background_color,
    now() - last_seen_at < interval '15 minutes' AS is_online,
    extract(year from age(birthday))::integer as "age",
    short_users.id AS invited_by_id,
    short_users.name AS invited_by_name,
    short_users.show_name AS invited_by_show_name,
    short_users.is_online AS invited_by_is_online,
    short_users.avatar AS invited_by_avatar
FROM mindwell.users
INNER JOIN mindwell.gender ON users.gender = gender.id
INNER JOIN mindwell.user_privacy ON users.privacy = user_privacy.id
INNER JOIN mindwell.font_family ON users.font_family = font_family.id
INNER JOIN mindwell.alignment ON users.text_alignment = alignment.id
LEFT JOIN mindwell.short_users ON users.invited_by = short_users.id;

CREATE VIEW mindwell.feed AS
SELECT entries.id, entries.created_at, rating, up_votes, down_votes,
    entries.title, cut_title, content, cut_content, edit_content, 
    has_cut, word_count,
    entry_privacy.type AS entry_privacy,
    is_votable, entries.comments_count,
    long_users.id AS author_id,
    long_users.name AS author_name, 
    long_users.show_name AS author_show_name,
    long_users.is_online AS author_is_online,
    long_users.avatar AS author_avatar,
    long_users.privacy AS author_privacy
FROM mindwell.long_users, mindwell.entries, mindwell.entry_privacy
WHERE long_users.id = entries.author_id 
    AND entry_privacy.id = entries.visible_for;

ALTER TABLE users 
ALTER COLUMN invited_by DROP NOT NULL;

INSERT INTO "mindwell"."notification_type" VALUES(6, 'invited');

INSERT INTO "mindwell"."user_privacy" VALUES(2, 'invited');
