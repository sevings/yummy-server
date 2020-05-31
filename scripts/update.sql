INSERT INTO "mindwell"."notification_type" VALUES(9, 'info');

-- CREATE TABLE "info" -----------------------------------------
CREATE TABLE "mindwell"."info" (
    "id" Serial NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "content" Text NOT NULL,
    "link" Text NOT NULL,
	CONSTRAINT "unique_info_id" PRIMARY KEY("id") );
;
-- -------------------------------------------------------------

INSERT INTO mindwell.info(content, link)
VALUES ('Пожалуйста, подтверди адрес свой почты. Теперь это обязательно.', '/account/email');

WITH subj AS (
    SELECT MAX(id) AS id
    FROM info
)
INSERT INTO mindwell.notifications(user_id, type, subject_id)
SELECT users.id, 9, subj.id
FROM users, subj
WHERE NOT verified;
