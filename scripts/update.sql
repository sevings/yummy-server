
-- CREATE TABLE "chats" ----------------------------------------
CREATE TABLE "mindwell"."chats" (
	"id" Serial NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	"creator_id" Integer NOT NULL,
	"partner_id" Integer NOT NULL,
	"last_message" Integer,
	CONSTRAINT "unique_chat_id" PRIMARY KEY( "id" ),
	CONSTRAINT "chat_creator" FOREIGN KEY ("creator_id") REFERENCES "mindwell"."users"("id") ON DELETE CASCADE,
	CONSTRAINT "chat_partner" FOREIGN KEY ("partner_id") REFERENCES "mindwell"."users"("id") ON DELETE CASCADE,
	CONSTRAINT "unique_chat_partners" UNIQUE ( "creator_id", "partner_id" ) );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_chat_id" --------------------------------
CREATE INDEX "index_chat_id" ON "mindwell"."chats" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_chat_creator_id" ------------------------
CREATE INDEX "index_chat_creator_id" ON "mindwell"."chats" USING btree( "creator_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_chat_partner_id" ------------------------
CREATE INDEX "index_chat_partner_id" ON "mindwell"."chats" USING btree( "partner_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_last_message_id" ------------------------
CREATE INDEX "index_last_message_id" ON "mindwell"."chats" USING btree( "last_message" );
-- -------------------------------------------------------------



-- CREATE TABLE "messages" -------------------------------------
CREATE TABLE "mindwell"."messages" (
	"id" Serial NOT NULL,
	"chat_id" Integer NOT NULL,
	"author_id" Integer NOT NULL,
	"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	"content" Text NOT NULL,
    "edit_content" Text NOT NULL,
	CONSTRAINT "unique_message_id" PRIMARY KEY( "id" ),
    CONSTRAINT "message_user_id" FOREIGN KEY("author_id") REFERENCES "mindwell"."users"("id") ON DELETE CASCADE,
    CONSTRAINT "message_chat_id" FOREIGN KEY("chat_id") REFERENCES "mindwell"."chats"("id") ON DELETE CASCADE );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_message_id" -----------------------------
CREATE INDEX "index_message_id" ON "mindwell"."messages" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_message_chat" ---------------------------
CREATE INDEX "index_message_chat" ON "mindwell"."messages" USING btree( "chat_id" );
-- -------------------------------------------------------------

ALTER TABLE "mindwell"."chats"
ADD CONSTRAINT "chat_last_message_id" FOREIGN KEY("last_message") REFERENCES "mindwell"."messages"("id");

CREATE OR REPLACE FUNCTION mindwell.set_last_message_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.chats
        SET last_message = NEW.id
        WHERE id = NEW.chat_id;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER last_messages_ins
    AFTER INSERT ON mindwell.messages
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_last_message_ins();

CREATE OR REPLACE FUNCTION mindwell.set_last_message_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.chats
        SET last_message = (
            SELECT max(messages.id)
            FROM messages
            WHERE chat_id = OLD.chat_id AND id <> OLD.id
        )
        WHERE last_message = OLD.id;

        RETURN OLD;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER last_messages_del
    BEFORE DELETE ON mindwell.messages
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_last_message_del();



-- CREATE TABLE "talkers" --------------------------------------
CREATE TABLE "mindwell"."talkers" (
	"chat_id" Integer NOT NULL,
	"user_id" Integer NOT NULL,
	"last_read" Integer,
	"unread_count" Integer DEFAULT 0 NOT NULL,
	"can_send" Boolean DEFAULT TRUE NOT NULL,
	CONSTRAINT "unique_talker_chat" PRIMARY KEY( "chat_id", "user_id" ),
    CONSTRAINT "talkers_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "talkers_chat_id" FOREIGN KEY("chat_id") REFERENCES "mindwell"."chats"("id"));
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_talkers_chat" ---------------------------
CREATE INDEX "index_talkers_chat" ON "mindwell"."talkers" USING btree( "chat_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_talkers_user" ---------------------------
CREATE INDEX "index_talkers_user" ON "mindwell"."talkers" USING btree( "user_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.count_unread_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.talkers
        SET unread_count = unread_count + 1
        WHERE talkers.chat_id = NEW.chat_id AND talkers.user_id <> NEW.author_id;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_unread_ins
    AFTER INSERT ON mindwell.messages
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_unread_ins();

CREATE OR REPLACE FUNCTION mindwell.count_unread_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.talkers
        SET unread_count = unread_count - 1
        WHERE talkers.chat_id = OLD.chat_id AND talkers.user_id <> OLD.author_id
            AND (last_read IS NULL OR last_read < OLD.id);

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_unread_del
    AFTER DELETE ON mindwell.messages
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_unread_del();



CREATE OR REPLACE FUNCTION mindwell.is_partner_ignoring(user_id INTEGER, chat_id INTEGER) RETURNS BOOLEAN AS $$
    BEGIN
        RETURN COALESCE((
                SELECT relations.type = (SELECT id FROM relation WHERE type = 'ignored')
                FROM relations
                WHERE to_id = user_id AND from_id = (
                        SELECT (CASE creator_id WHEN user_id THEN partner_id ELSE creator_id END)
                        FROM chats
                        WHERE id = chat_id
                    )
            ), FALSE);
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.is_invited(user_id INTEGER) RETURNS BOOLEAN AS $$
    BEGIN
        RETURN (SELECT invited_by IS NOT NULL FROM users WHERE id = user_id);
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.set_can_send_ins() RETURNS TRIGGER AS $$
    BEGIN
        NEW.can_send = (
            SELECT is_invited(NEW.user_id) AND NOT is_partner_ignoring(NEW.user_id, NEW.chat_id)
        );

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER can_send_ins
    BEFORE INSERT ON mindwell.talkers
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_can_send_ins();

CREATE OR REPLACE FUNCTION mindwell.set_can_send_invited() RETURNS TRIGGER AS $$
    BEGIN
        IF NEW.invited_by <> OLD.invited_by THEN
            UPDATE mindwell.talkers
            SET can_send = NOT is_partner_ignoring(user_id, chat_id)
            WHERE user_id = NEW.id;
        END IF;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER can_send_invited
    AFTER UPDATE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_can_send_invited();

CREATE OR REPLACE FUNCTION mindwell.set_can_send_related() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.talkers
        SET can_send = (NEW.type != (SELECT id FROM relation WHERE type = 'ignored') AND is_invited(NEW.to_id))
        WHERE chat_id = (
            SELECT id
            FROM chats
            WHERE (creator_id = NEW.to_id AND partner_id = NEW.from_id)
                OR (creator_id = NEW.from_id AND partner_id = NEW.to_id)
        );

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER can_send_related
    AFTER UPDATE ON mindwell.relations
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_can_send_related();
