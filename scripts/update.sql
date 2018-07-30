CREATE UNIQUE INDEX "index_invite_words" ON "mindwell"."invites" USING btree( "word1", "word2", "word3" );
