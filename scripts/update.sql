ALTER TABLE users
ADD COLUMN "telegram_comments" Boolean NOT NULL DEFAULT TRUE;

ALTER TABLE users
ADD COLUMN "telegram_followers" Boolean NOT NULL DEFAULT TRUE;

ALTER TABLE users
ADD COLUMN "telegram_invites" Boolean NOT NULL DEFAULT TRUE;

ALTER TABLE users
ADD COLUMN "telegram_messages" Boolean NOT NULL DEFAULT TRUE;
