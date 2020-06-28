ALTER TABLE talkers
DROP CONSTRAINT talkers_user_id;

ALTER TABLE talkers
ADD CONSTRAINT "talkers_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id") ON DELETE CASCADE;

ALTER TABLE talkers
DROP CONSTRAINT talkers_chat_id;

ALTER TABLE talkers
ADD CONSTRAINT "talkers_chat_id" FOREIGN KEY("chat_id") REFERENCES "mindwell"."chats"("id") ON DELETE CASCADE;
