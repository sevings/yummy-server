-- noinspection SqlNoDataSourceInspectionForFile

CREATE SCHEMA "mindwell";

CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER DATABASE mindwell SET search_path TO mindwell, public;

-- CREATE TABLE "gender" ---------------------------------
CREATE TABLE "mindwell"."gender" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."gender" VALUES(0, 'not set');
INSERT INTO "mindwell"."gender" VALUES(1, 'male');
INSERT INTO "mindwell"."gender" VALUES(2, 'female');
-- -------------------------------------------------------------



-- CREATE TABLE "user_privacy" ---------------------------------
CREATE TABLE "mindwell"."user_privacy" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."user_privacy" VALUES(0, 'all');
INSERT INTO "mindwell"."user_privacy" VALUES(1, 'followers');
INSERT INTO "mindwell"."user_privacy" VALUES(2, 'invited');
-- -------------------------------------------------------------



-- CREATE TYPE "font_family" -----------------------------------
CREATE TABLE "mindwell"."font_family" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."font_family" VALUES(0, 'Arial');
-- -------------------------------------------------------------



-- CREATE TYPE "alignment" -------------------------------------
CREATE TABLE "mindwell"."alignment" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."alignment" VALUES(0, 'left');
INSERT INTO "mindwell"."alignment" VALUES(1, 'right');
INSERT INTO "mindwell"."alignment" VALUES(2, 'center');
INSERT INTO "mindwell"."alignment" VALUES(3, 'justify');
-- -------------------------------------------------------------


CREATE OR REPLACE FUNCTION next_user_rank() RETURNS INTEGER AS $$
    DECLARE
        pos INTEGER;
    BEGIN
        pos = (
            SELECT COUNT(*) + 1
            FROM users
            WHERE karma >= 0
        );
        RETURN pos;
    END;
$$ language plpgsql;

CREATE OR REPLACE FUNCTION to_search_string(name TEXT, show_name TEXT, country TEXT, city TEXT)
   RETURNS TEXT AS $$
BEGIN
  RETURN name || ' ' || show_name || ' ' || country || ' ' || city;
END
$$ LANGUAGE plpgsql IMMUTABLE;

-- CREATE TABLE "users" ----------------------------------------
CREATE TABLE "mindwell"."users" (
	"id" Serial NOT NULL,
	"name" Text NOT NULL,
	"show_name" Text DEFAULT '' NOT NULL,
	"password_hash" Bytea NOT NULL,
	"gender" Integer DEFAULT 0 NOT NULL,
	"is_daylog" Boolean DEFAULT false NOT NULL,
    "show_in_tops" Boolean DEFAULT true NOT NULL,
	"privacy" Integer DEFAULT 0 NOT NULL,
	"title" Text DEFAULT '' NOT NULL,
	"last_seen_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "rank" Integer DEFAULT next_user_rank() NOT NULL,
	"karma" Real DEFAULT 0 NOT NULL,
	"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	"last_invite" Date DEFAULT CURRENT_DATE NOT NULL,
    "invited_by" Integer,
	"birthday" Date,
	"css" Text DEFAULT '' NOT NULL,
	"entries_count" Integer DEFAULT 0 NOT NULL,
	"followings_count" Integer DEFAULT 0 NOT NULL,
	"followers_count" Integer DEFAULT 0 NOT NULL,
	"comments_count" Integer DEFAULT 0 NOT NULL,
	"ignored_count" Integer DEFAULT 0 NOT NULL,
	"invited_count" Integer DEFAULT 0 NOT NULL,
	"favorites_count" Integer DEFAULT 0 NOT NULL,
	"tags_count" Integer DEFAULT 0 NOT NULL,
	"country" Text DEFAULT '' NOT NULL,
	"city" Text DEFAULT '' NOT NULL,
	"email" Text NOT NULL,
	"verified" Boolean DEFAULT false NOT NULL,
    "api_key" Text NOT NULL,
    "valid_thru" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP + interval '6 months' NOT NULL,
	"avatar" Text DEFAULT '' NOT NULL,
	"cover" Text DEFAULT '' NOT NULL,
	"font_family" Integer DEFAULT 0 NOT NULL,
	"font_size" SmallInt DEFAULT 100 NOT NULL,
	"text_alignment" Integer DEFAULT 0 NOT NULL,
	"text_color" Character( 7 ) DEFAULT '#000000' NOT NULL,
	"background_color" Character( 7 ) DEFAULT '#ffffff' NOT NULL,
    "email_comments" Boolean NOT NULL DEFAULT FALSE,
    "email_followers" Boolean NOT NULL DEFAULT FALSE,
    "email_invites" Boolean NOT NULL DEFAULT FALSE,
    "invite_ban" Date DEFAULT CURRENT_DATE NOT NULL,
    "vote_ban" Date DEFAULT CURRENT_DATE + interval '1 month' NOT NULL,
    "comment_ban" Date DEFAULT CURRENT_DATE NOT NULL,
    "live_ban" Date DEFAULT CURRENT_DATE NOT NULL,
    "adm_ban" Boolean DEFAULT TRUE NOT NULL,
    "telegram" Integer,
	CONSTRAINT "unique_user_id" PRIMARY KEY( "id" ),
    CONSTRAINT "enum_user_gender" FOREIGN KEY("gender") REFERENCES "mindwell"."gender"("id"),
    CONSTRAINT "enum_user_privacy" FOREIGN KEY("privacy") REFERENCES "mindwell"."user_privacy"("id"),
    CONSTRAINT "enum_user_alignment" FOREIGN KEY("text_alignment") REFERENCES "mindwell"."alignment"("id"),
    CONSTRAINT "enum_user_font_family" FOREIGN KEY("font_family") REFERENCES "mindwell"."font_family"("id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_user_id" --------------------------------
CREATE UNIQUE INDEX "index_user_id" ON "mindwell"."users" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_user_name" ------------------------------
CREATE UNIQUE INDEX "index_user_name" ON "mindwell"."users" USING btree( lower("name") );
-- -------------------------------------------------------------

-- CREATE INDEX "index_user_email" -----------------------------
CREATE UNIQUE INDEX "index_user_email" ON "mindwell"."users" USING btree( lower("email") );
-- -------------------------------------------------------------

-- CREATE INDEX "index_token_user" -----------------------------
CREATE UNIQUE INDEX "index_user_key" ON "mindwell"."users" USING btree( "api_key" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_telegram" -------------------------------
CREATE UNIQUE INDEX "index_telegram" ON "mindwell"."users" USING btree( "telegram" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_invited_by" -----------------------------
CREATE INDEX "index_invited_by" ON "mindwell"."users" USING btree( "invited_by" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_user_rank" ------------------------------
CREATE INDEX "index_user_rank" ON "mindwell"."users" USING btree( "rank" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_user_search" ----------------------------
CREATE INDEX "index_user_search" ON "mindwell"."users" USING GIST 
    (to_search_string("name", "show_name", "country", "city") gist_trgm_ops);
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.count_invited_upd() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD.invited_by = NEW.invited_by) THEN
            RETURN NULL;
        END IF;

        UPDATE mindwell.users
        SET invited_count = invited_count + 1
        WHERE id = NEW.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_upd
    AFTER UPDATE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_invited_upd();

CREATE OR REPLACE FUNCTION mindwell.count_invited_del() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD.invited_by IS NULL) THEN
            RETURN NULL;
        END IF;

        UPDATE mindwell.users
        SET invited_count = invited_count - 1
        WHERE id = OLD.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited_del
    AFTER DELETE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_invited_del();

CREATE OR REPLACE FUNCTION mindwell.allow_adm_upd() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD.invited_by <> NEW.invited_by) THEN
            NEW.adm_ban = false;
        END IF;

        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER alw_adm_upd
    BEFORE UPDATE ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.allow_adm_upd();

    

-- CREATE TABLE "adm" ------------------------------------------
CREATE TABLE "mindwell"."adm" (
	"name" Text NOT NULL,
    "fullname" Text NOT NULL,
    "postcode" Text NOT NULL,
    "country" Text NOT NULL,
    "address" Text NOT NULL,
    "comment" Text NOT NULL,
    "anonymous" Boolean NOT NULL,
    "grandfather" Text NOT NULL DEFAULT '',
    "sent" Boolean NOT NULL DEFAULT false,
    "received" Boolean NOT NULL DEFAULT false );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_adm" ------------------------------------
CREATE UNIQUE INDEX "index_adm" ON "mindwell"."adm" USING btree( lower("name") );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.ban_adm() RETURNS VOID AS $$
    UPDATE users 
    SET adm_ban = true
    WHERE name IN (
        SELECT gs.name 
        FROM adm AS gs
        JOIN adm AS gf ON gf.grandfather = gs.name
        WHERE NOT gf.sent OR (gs.sent AND NOT gs.received)
    );
$$ LANGUAGE SQL;

    

-- CREATE TABLE "invite_words" ---------------------------------
CREATE TABLE "mindwell"."invite_words" (
    "id" Serial NOT NULL,
    "word" Text NOT NULL,
	CONSTRAINT "unique_word_id" PRIMARY KEY( "id" ),
	CONSTRAINT "unique_word" UNIQUE( "word" ) );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_invite_word_id" -------------------------
CREATE UNIQUE INDEX "index_invite_word_id" ON "mindwell"."invite_words" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_invite_word" ----------------------------
CREATE UNIQUE INDEX "index_invite_word" ON "mindwell"."invite_words" USING btree( "word" );
-- -------------------------------------------------------------

INSERT INTO mindwell.invite_words ("word") VALUES('acknown');
INSERT INTO mindwell.invite_words ("word") VALUES('aery');
INSERT INTO mindwell.invite_words ("word") VALUES('affectioned');
INSERT INTO mindwell.invite_words ("word") VALUES('agnize');
INSERT INTO mindwell.invite_words ("word") VALUES('ambition');
INSERT INTO mindwell.invite_words ("word") VALUES('amerce');
INSERT INTO mindwell.invite_words ("word") VALUES('anters');
INSERT INTO mindwell.invite_words ("word") VALUES('argal');
INSERT INTO mindwell.invite_words ("word") VALUES('arrant');
INSERT INTO mindwell.invite_words ("word") VALUES('arras');
INSERT INTO mindwell.invite_words ("word") VALUES('asquint');
INSERT INTO mindwell.invite_words ("word") VALUES('atomies');
INSERT INTO mindwell.invite_words ("word") VALUES('augurers');
INSERT INTO mindwell.invite_words ("word") VALUES('bastinado');
INSERT INTO mindwell.invite_words ("word") VALUES('batten');
INSERT INTO mindwell.invite_words ("word") VALUES('bawbling');
INSERT INTO mindwell.invite_words ("word") VALUES('bawcock');
INSERT INTO mindwell.invite_words ("word") VALUES('bawd');
INSERT INTO mindwell.invite_words ("word") VALUES('behoveful');
INSERT INTO mindwell.invite_words ("word") VALUES('beldams');
INSERT INTO mindwell.invite_words ("word") VALUES('belike');
INSERT INTO mindwell.invite_words ("word") VALUES('berattle');
INSERT INTO mindwell.invite_words ("word") VALUES('beshrew');
INSERT INTO mindwell.invite_words ("word") VALUES('betid');
INSERT INTO mindwell.invite_words ("word") VALUES('betimes');
INSERT INTO mindwell.invite_words ("word") VALUES('betoken');
INSERT INTO mindwell.invite_words ("word") VALUES('bewray');
INSERT INTO mindwell.invite_words ("word") VALUES('biddy');
INSERT INTO mindwell.invite_words ("word") VALUES('bilboes');
INSERT INTO mindwell.invite_words ("word") VALUES('blasted');
INSERT INTO mindwell.invite_words ("word") VALUES('blazon');
INSERT INTO mindwell.invite_words ("word") VALUES('bodements');
INSERT INTO mindwell.invite_words ("word") VALUES('bodkin');
INSERT INTO mindwell.invite_words ("word") VALUES('bombard');
INSERT INTO mindwell.invite_words ("word") VALUES('bootless');
INSERT INTO mindwell.invite_words ("word") VALUES('bosky');
INSERT INTO mindwell.invite_words ("word") VALUES('bowers');
INSERT INTO mindwell.invite_words ("word") VALUES('brach');
INSERT INTO mindwell.invite_words ("word") VALUES('brainsickly');
INSERT INTO mindwell.invite_words ("word") VALUES('brock');
INSERT INTO mindwell.invite_words ("word") VALUES('bruit');
INSERT INTO mindwell.invite_words ("word") VALUES('buckler');
INSERT INTO mindwell.invite_words ("word") VALUES('busky');
INSERT INTO mindwell.invite_words ("word") VALUES('caitiff');
INSERT INTO mindwell.invite_words ("word") VALUES('caliver');
INSERT INTO mindwell.invite_words ("word") VALUES('callet');
INSERT INTO mindwell.invite_words ("word") VALUES('cantons');
INSERT INTO mindwell.invite_words ("word") VALUES('carded');
INSERT INTO mindwell.invite_words ("word") VALUES('carrions');
INSERT INTO mindwell.invite_words ("word") VALUES('cashiered');
INSERT INTO mindwell.invite_words ("word") VALUES('casing');
INSERT INTO mindwell.invite_words ("word") VALUES('catch');
INSERT INTO mindwell.invite_words ("word") VALUES('caterwauling');
INSERT INTO mindwell.invite_words ("word") VALUES('cautel');
INSERT INTO mindwell.invite_words ("word") VALUES('cerecloth');
INSERT INTO mindwell.invite_words ("word") VALUES('cerements');
INSERT INTO mindwell.invite_words ("word") VALUES('certes');
INSERT INTO mindwell.invite_words ("word") VALUES('champain');
INSERT INTO mindwell.invite_words ("word") VALUES('chaps');
INSERT INTO mindwell.invite_words ("word") VALUES('charactery');
INSERT INTO mindwell.invite_words ("word") VALUES('chariest');
INSERT INTO mindwell.invite_words ("word") VALUES('charmingly');
INSERT INTO mindwell.invite_words ("word") VALUES('chinks');
INSERT INTO mindwell.invite_words ("word") VALUES('chopt');
INSERT INTO mindwell.invite_words ("word") VALUES('chough');
INSERT INTO mindwell.invite_words ("word") VALUES('civet');
INSERT INTO mindwell.invite_words ("word") VALUES('clepe');
INSERT INTO mindwell.invite_words ("word") VALUES('climatures');
INSERT INTO mindwell.invite_words ("word") VALUES('clodpole');
INSERT INTO mindwell.invite_words ("word") VALUES('cobbler');
INSERT INTO mindwell.invite_words ("word") VALUES('cockatrices');
INSERT INTO mindwell.invite_words ("word") VALUES('collied');
INSERT INTO mindwell.invite_words ("word") VALUES('collier');
INSERT INTO mindwell.invite_words ("word") VALUES('colour');
INSERT INTO mindwell.invite_words ("word") VALUES('compass');
INSERT INTO mindwell.invite_words ("word") VALUES('comptible');
INSERT INTO mindwell.invite_words ("word") VALUES('conceit');
INSERT INTO mindwell.invite_words ("word") VALUES('condition');
INSERT INTO mindwell.invite_words ("word") VALUES('continuate');
INSERT INTO mindwell.invite_words ("word") VALUES('corky');
INSERT INTO mindwell.invite_words ("word") VALUES('coronets');
INSERT INTO mindwell.invite_words ("word") VALUES('corse');
INSERT INTO mindwell.invite_words ("word") VALUES('coxcomb');
INSERT INTO mindwell.invite_words ("word") VALUES('coystrill');
INSERT INTO mindwell.invite_words ("word") VALUES('cozen');
INSERT INTO mindwell.invite_words ("word") VALUES('cozier');
INSERT INTO mindwell.invite_words ("word") VALUES('crisped');
INSERT INTO mindwell.invite_words ("word") VALUES('crochets');
INSERT INTO mindwell.invite_words ("word") VALUES('crossed');
INSERT INTO mindwell.invite_words ("word") VALUES('crowner');
INSERT INTO mindwell.invite_words ("word") VALUES('cubiculo');
INSERT INTO mindwell.invite_words ("word") VALUES('cursy');
INSERT INTO mindwell.invite_words ("word") VALUES('dallying');
INSERT INTO mindwell.invite_words ("word") VALUES('dateless');
INSERT INTO mindwell.invite_words ("word") VALUES('daws');
INSERT INTO mindwell.invite_words ("word") VALUES('denotement');
INSERT INTO mindwell.invite_words ("word") VALUES('dilate');
INSERT INTO mindwell.invite_words ("word") VALUES('dissemble');
INSERT INTO mindwell.invite_words ("word") VALUES('distaff');
INSERT INTO mindwell.invite_words ("word") VALUES('distemperature');
INSERT INTO mindwell.invite_words ("word") VALUES('doit');
INSERT INTO mindwell.invite_words ("word") VALUES('doublet');
INSERT INTO mindwell.invite_words ("word") VALUES('doves');
INSERT INTO mindwell.invite_words ("word") VALUES('drabbing');
INSERT INTO mindwell.invite_words ("word") VALUES('dram');
INSERT INTO mindwell.invite_words ("word") VALUES('drossy');
INSERT INTO mindwell.invite_words ("word") VALUES('dudgeon');
INSERT INTO mindwell.invite_words ("word") VALUES('dunnest');
INSERT INTO mindwell.invite_words ("word") VALUES('eanlings');
INSERT INTO mindwell.invite_words ("word") VALUES('elflocks');
INSERT INTO mindwell.invite_words ("word") VALUES('eliads');
INSERT INTO mindwell.invite_words ("word") VALUES('encave');
INSERT INTO mindwell.invite_words ("word") VALUES('enchafed');
INSERT INTO mindwell.invite_words ("word") VALUES('endues');
INSERT INTO mindwell.invite_words ("word") VALUES('engluts');
INSERT INTO mindwell.invite_words ("word") VALUES('ensteeped');
INSERT INTO mindwell.invite_words ("word") VALUES('envy');
INSERT INTO mindwell.invite_words ("word") VALUES('enwheel');
INSERT INTO mindwell.invite_words ("word") VALUES('erns');
INSERT INTO mindwell.invite_words ("word") VALUES('extremities');
INSERT INTO mindwell.invite_words ("word") VALUES('eyeless');
INSERT INTO mindwell.invite_words ("word") VALUES('fable');
INSERT INTO mindwell.invite_words ("word") VALUES('factious');
INSERT INTO mindwell.invite_words ("word") VALUES('fadge');
INSERT INTO mindwell.invite_words ("word") VALUES('fain');
INSERT INTO mindwell.invite_words ("word") VALUES('fashion');
INSERT INTO mindwell.invite_words ("word") VALUES('favour');
INSERT INTO mindwell.invite_words ("word") VALUES('festinate');
INSERT INTO mindwell.invite_words ("word") VALUES('fetches');
INSERT INTO mindwell.invite_words ("word") VALUES('figures');
INSERT INTO mindwell.invite_words ("word") VALUES('fleer');
INSERT INTO mindwell.invite_words ("word") VALUES('fleering');
INSERT INTO mindwell.invite_words ("word") VALUES('flote');
INSERT INTO mindwell.invite_words ("word") VALUES('flowerets');
INSERT INTO mindwell.invite_words ("word") VALUES('fobbed');
INSERT INTO mindwell.invite_words ("word") VALUES('foison');
INSERT INTO mindwell.invite_words ("word") VALUES('fopped');
INSERT INTO mindwell.invite_words ("word") VALUES('fordid');
INSERT INTO mindwell.invite_words ("word") VALUES('forks');
INSERT INTO mindwell.invite_words ("word") VALUES('franklin');
INSERT INTO mindwell.invite_words ("word") VALUES('frieze');
INSERT INTO mindwell.invite_words ("word") VALUES('frippery');
INSERT INTO mindwell.invite_words ("word") VALUES('fulsome');
INSERT INTO mindwell.invite_words ("word") VALUES('fust');
INSERT INTO mindwell.invite_words ("word") VALUES('fustian');
INSERT INTO mindwell.invite_words ("word") VALUES('gage');
INSERT INTO mindwell.invite_words ("word") VALUES('gaged');
INSERT INTO mindwell.invite_words ("word") VALUES('gallow');
INSERT INTO mindwell.invite_words ("word") VALUES('gamesome');
INSERT INTO mindwell.invite_words ("word") VALUES('gaskins');
INSERT INTO mindwell.invite_words ("word") VALUES('gasted');
INSERT INTO mindwell.invite_words ("word") VALUES('gauntlet');
INSERT INTO mindwell.invite_words ("word") VALUES('gentle');
INSERT INTO mindwell.invite_words ("word") VALUES('glazed');
INSERT INTO mindwell.invite_words ("word") VALUES('gleek');
INSERT INTO mindwell.invite_words ("word") VALUES('goatish');
INSERT INTO mindwell.invite_words ("word") VALUES('goodyears');
INSERT INTO mindwell.invite_words ("word") VALUES('goose');
INSERT INTO mindwell.invite_words ("word") VALUES('gouts');
INSERT INTO mindwell.invite_words ("word") VALUES('gramercy');
INSERT INTO mindwell.invite_words ("word") VALUES('grise');
INSERT INTO mindwell.invite_words ("word") VALUES('grizzled');
INSERT INTO mindwell.invite_words ("word") VALUES('groundings');
INSERT INTO mindwell.invite_words ("word") VALUES('gudgeon');
INSERT INTO mindwell.invite_words ("word") VALUES('gull');
INSERT INTO mindwell.invite_words ("word") VALUES('guttered');
INSERT INTO mindwell.invite_words ("word") VALUES('hams');
INSERT INTO mindwell.invite_words ("word") VALUES('haply');
INSERT INTO mindwell.invite_words ("word") VALUES('hardiment');
INSERT INTO mindwell.invite_words ("word") VALUES('harpy');
INSERT INTO mindwell.invite_words ("word") VALUES('hart');
INSERT INTO mindwell.invite_words ("word") VALUES('heath');
INSERT INTO mindwell.invite_words ("word") VALUES('hests');
INSERT INTO mindwell.invite_words ("word") VALUES('hilding');
INSERT INTO mindwell.invite_words ("word") VALUES('hinds');
INSERT INTO mindwell.invite_words ("word") VALUES('holidam');
INSERT INTO mindwell.invite_words ("word") VALUES('holp');
INSERT INTO mindwell.invite_words ("word") VALUES('housewives');
INSERT INTO mindwell.invite_words ("word") VALUES('humour');
INSERT INTO mindwell.invite_words ("word") VALUES('hurlyburly');
INSERT INTO mindwell.invite_words ("word") VALUES('husbandry');
INSERT INTO mindwell.invite_words ("word") VALUES('ides');
INSERT INTO mindwell.invite_words ("word") VALUES('import');
INSERT INTO mindwell.invite_words ("word") VALUES('incarnadine');
INSERT INTO mindwell.invite_words ("word") VALUES('indign');
INSERT INTO mindwell.invite_words ("word") VALUES('ingraft');
INSERT INTO mindwell.invite_words ("word") VALUES('ingrafted');
INSERT INTO mindwell.invite_words ("word") VALUES('insuppressive');
INSERT INTO mindwell.invite_words ("word") VALUES('intentively');
INSERT INTO mindwell.invite_words ("word") VALUES('intermit');
INSERT INTO mindwell.invite_words ("word") VALUES('jaunce');
INSERT INTO mindwell.invite_words ("word") VALUES('jaundice');
INSERT INTO mindwell.invite_words ("word") VALUES('jealous');
INSERT INTO mindwell.invite_words ("word") VALUES('jointress');
INSERT INTO mindwell.invite_words ("word") VALUES('jowls');
INSERT INTO mindwell.invite_words ("word") VALUES('knapped');
INSERT INTO mindwell.invite_words ("word") VALUES('ladybird');
INSERT INTO mindwell.invite_words ("word") VALUES('leasing');
INSERT INTO mindwell.invite_words ("word") VALUES('leman');
INSERT INTO mindwell.invite_words ("word") VALUES('lethe');
INSERT INTO mindwell.invite_words ("word") VALUES('lief');
INSERT INTO mindwell.invite_words ("word") VALUES('liver');
INSERT INTO mindwell.invite_words ("word") VALUES('livings');
INSERT INTO mindwell.invite_words ("word") VALUES('loath');
INSERT INTO mindwell.invite_words ("word") VALUES('loggerheads');
INSERT INTO mindwell.invite_words ("word") VALUES('lown');
INSERT INTO mindwell.invite_words ("word") VALUES('magnificoes');
INSERT INTO mindwell.invite_words ("word") VALUES('maidenhead');
INSERT INTO mindwell.invite_words ("word") VALUES('malapert');
INSERT INTO mindwell.invite_words ("word") VALUES('marchpane');
INSERT INTO mindwell.invite_words ("word") VALUES('marry');
INSERT INTO mindwell.invite_words ("word") VALUES('masterless');
INSERT INTO mindwell.invite_words ("word") VALUES('maugre');
INSERT INTO mindwell.invite_words ("word") VALUES('mazzard');
INSERT INTO mindwell.invite_words ("word") VALUES('meet');
INSERT INTO mindwell.invite_words ("word") VALUES('meetest');
INSERT INTO mindwell.invite_words ("word") VALUES('meiny');
INSERT INTO mindwell.invite_words ("word") VALUES('meshes');
INSERT INTO mindwell.invite_words ("word") VALUES('micher');
INSERT INTO mindwell.invite_words ("word") VALUES('minion');
INSERT INTO mindwell.invite_words ("word") VALUES('misprision');
INSERT INTO mindwell.invite_words ("word") VALUES('moo');
INSERT INTO mindwell.invite_words ("word") VALUES('mooncalf');
INSERT INTO mindwell.invite_words ("word") VALUES('mountebanks');
INSERT INTO mindwell.invite_words ("word") VALUES('mushrumps');
INSERT INTO mindwell.invite_words ("word") VALUES('mute');
INSERT INTO mindwell.invite_words ("word") VALUES('naughty');
INSERT INTO mindwell.invite_words ("word") VALUES('nonce');
INSERT INTO mindwell.invite_words ("word") VALUES('nuncle');
INSERT INTO mindwell.invite_words ("word") VALUES('occulted');
INSERT INTO mindwell.invite_words ("word") VALUES('ordinary');
INSERT INTO mindwell.invite_words ("word") VALUES('othergates');
INSERT INTO mindwell.invite_words ("word") VALUES('overname');
INSERT INTO mindwell.invite_words ("word") VALUES('paddock');
INSERT INTO mindwell.invite_words ("word") VALUES('palmy');
INSERT INTO mindwell.invite_words ("word") VALUES('palter');
INSERT INTO mindwell.invite_words ("word") VALUES('parle');
INSERT INTO mindwell.invite_words ("word") VALUES('patch');
INSERT INTO mindwell.invite_words ("word") VALUES('paunch');
INSERT INTO mindwell.invite_words ("word") VALUES('pearl');
INSERT INTO mindwell.invite_words ("word") VALUES('peize');
INSERT INTO mindwell.invite_words ("word") VALUES('pennyworths');
INSERT INTO mindwell.invite_words ("word") VALUES('perdy');
INSERT INTO mindwell.invite_words ("word") VALUES('pignuts');
INSERT INTO mindwell.invite_words ("word") VALUES('portance');
INSERT INTO mindwell.invite_words ("word") VALUES('possets');
INSERT INTO mindwell.invite_words ("word") VALUES('posy');
INSERT INTO mindwell.invite_words ("word") VALUES('praetor');
INSERT INTO mindwell.invite_words ("word") VALUES('prate');
INSERT INTO mindwell.invite_words ("word") VALUES('prick');
INSERT INTO mindwell.invite_words ("word") VALUES('primy');
INSERT INTO mindwell.invite_words ("word") VALUES('princox');
INSERT INTO mindwell.invite_words ("word") VALUES('prithee');
INSERT INTO mindwell.invite_words ("word") VALUES('prodigies');
INSERT INTO mindwell.invite_words ("word") VALUES('proper');
INSERT INTO mindwell.invite_words ("word") VALUES('prorogued');
INSERT INTO mindwell.invite_words ("word") VALUES('pudder');
INSERT INTO mindwell.invite_words ("word") VALUES('puddled');
INSERT INTO mindwell.invite_words ("word") VALUES('puling');
INSERT INTO mindwell.invite_words ("word") VALUES('purblind');
INSERT INTO mindwell.invite_words ("word") VALUES('pursy');
INSERT INTO mindwell.invite_words ("word") VALUES('quailing');
INSERT INTO mindwell.invite_words ("word") VALUES('quaint');
INSERT INTO mindwell.invite_words ("word") VALUES('quiddities');
INSERT INTO mindwell.invite_words ("word") VALUES('quilets');
INSERT INTO mindwell.invite_words ("word") VALUES('quillets');
INSERT INTO mindwell.invite_words ("word") VALUES('reference');
INSERT INTO mindwell.invite_words ("word") VALUES('instrument');
INSERT INTO mindwell.invite_words ("word") VALUES('ranker');
INSERT INTO mindwell.invite_words ("word") VALUES('rated');
INSERT INTO mindwell.invite_words ("word") VALUES('razes');
INSERT INTO mindwell.invite_words ("word") VALUES('receiving');
INSERT INTO mindwell.invite_words ("word") VALUES('reechy');
INSERT INTO mindwell.invite_words ("word") VALUES('reeking');
INSERT INTO mindwell.invite_words ("word") VALUES('remembrances');
INSERT INTO mindwell.invite_words ("word") VALUES('rheumy');
INSERT INTO mindwell.invite_words ("word") VALUES('rive');
INSERT INTO mindwell.invite_words ("word") VALUES('robustious');
INSERT INTO mindwell.invite_words ("word") VALUES('romage');
INSERT INTO mindwell.invite_words ("word") VALUES('ronyon');
INSERT INTO mindwell.invite_words ("word") VALUES('rouse');
INSERT INTO mindwell.invite_words ("word") VALUES('sallies');
INSERT INTO mindwell.invite_words ("word") VALUES('saws');
INSERT INTO mindwell.invite_words ("word") VALUES('scanted');
INSERT INTO mindwell.invite_words ("word") VALUES('scarfed');
INSERT INTO mindwell.invite_words ("word") VALUES('scrimers');
INSERT INTO mindwell.invite_words ("word") VALUES('scutcheon');
INSERT INTO mindwell.invite_words ("word") VALUES('seel');
INSERT INTO mindwell.invite_words ("word") VALUES('sennet');
INSERT INTO mindwell.invite_words ("word") VALUES('sequestration');
INSERT INTO mindwell.invite_words ("word") VALUES('shent');
INSERT INTO mindwell.invite_words ("word") VALUES('shoon');
INSERT INTO mindwell.invite_words ("word") VALUES('shoughs');
INSERT INTO mindwell.invite_words ("word") VALUES('shrift');
INSERT INTO mindwell.invite_words ("word") VALUES('sleave');
INSERT INTO mindwell.invite_words ("word") VALUES('slubber');
INSERT INTO mindwell.invite_words ("word") VALUES('smilets');
INSERT INTO mindwell.invite_words ("word") VALUES('sonties');
INSERT INTO mindwell.invite_words ("word") VALUES('sooth');
INSERT INTO mindwell.invite_words ("word") VALUES('sounded');
INSERT INTO mindwell.invite_words ("word") VALUES('spleen');
INSERT INTO mindwell.invite_words ("word") VALUES('splenetive');
INSERT INTO mindwell.invite_words ("word") VALUES('spongy');
INSERT INTO mindwell.invite_words ("word") VALUES('springe');
INSERT INTO mindwell.invite_words ("word") VALUES('steads');
INSERT INTO mindwell.invite_words ("word") VALUES('still');
INSERT INTO mindwell.invite_words ("word") VALUES('stoup');
INSERT INTO mindwell.invite_words ("word") VALUES('stronds');
INSERT INTO mindwell.invite_words ("word") VALUES('suit');
INSERT INTO mindwell.invite_words ("word") VALUES('swoopstake');
INSERT INTO mindwell.invite_words ("word") VALUES('swounded');
INSERT INTO mindwell.invite_words ("word") VALUES('tabor');
INSERT INTO mindwell.invite_words ("word") VALUES('taper');
INSERT INTO mindwell.invite_words ("word") VALUES('teen');
INSERT INTO mindwell.invite_words ("word") VALUES('tenders');
INSERT INTO mindwell.invite_words ("word") VALUES('termagant');
INSERT INTO mindwell.invite_words ("word") VALUES('tetchy');
INSERT INTO mindwell.invite_words ("word") VALUES('tinkers');
INSERT INTO mindwell.invite_words ("word") VALUES('topgallant');
INSERT INTO mindwell.invite_words ("word") VALUES('traffic');
INSERT INTO mindwell.invite_words ("word") VALUES('traject');
INSERT INTO mindwell.invite_words ("word") VALUES('trencher');
INSERT INTO mindwell.invite_words ("word") VALUES('trimmed');
INSERT INTO mindwell.invite_words ("word") VALUES('tristful');
INSERT INTO mindwell.invite_words ("word") VALUES('trowest');
INSERT INTO mindwell.invite_words ("word") VALUES('truncheon');
INSERT INTO mindwell.invite_words ("word") VALUES('unbend');
INSERT INTO mindwell.invite_words ("word") VALUES('unbitted');
INSERT INTO mindwell.invite_words ("word") VALUES('unbound');
INSERT INTO mindwell.invite_words ("word") VALUES('unbraced');
INSERT INTO mindwell.invite_words ("word") VALUES('unbruised');
INSERT INTO mindwell.invite_words ("word") VALUES('undone');
INSERT INTO mindwell.invite_words ("word") VALUES('ungently');
INSERT INTO mindwell.invite_words ("word") VALUES('unhoused');
INSERT INTO mindwell.invite_words ("word") VALUES('unmake');
INSERT INTO mindwell.invite_words ("word") VALUES('unprevailing');
INSERT INTO mindwell.invite_words ("word") VALUES('unprovide');
INSERT INTO mindwell.invite_words ("word") VALUES('unreclaimed');
INSERT INTO mindwell.invite_words ("word") VALUES('unstuffed');
INSERT INTO mindwell.invite_words ("word") VALUES('untaught');
INSERT INTO mindwell.invite_words ("word") VALUES('untented');
INSERT INTO mindwell.invite_words ("word") VALUES('unthrifty');
INSERT INTO mindwell.invite_words ("word") VALUES('unyoke');
INSERT INTO mindwell.invite_words ("word") VALUES('usance');
INSERT INTO mindwell.invite_words ("word") VALUES('vailing');
INSERT INTO mindwell.invite_words ("word") VALUES('varlets');
INSERT INTO mindwell.invite_words ("word") VALUES('verdure');
INSERT INTO mindwell.invite_words ("word") VALUES('villanies');
INSERT INTO mindwell.invite_words ("word") VALUES('vizards');
INSERT INTO mindwell.invite_words ("word") VALUES('wafter');
INSERT INTO mindwell.invite_words ("word") VALUES('welkin');
INSERT INTO mindwell.invite_words ("word") VALUES('weraday');
INSERT INTO mindwell.invite_words ("word") VALUES('whoreson');
INSERT INTO mindwell.invite_words ("word") VALUES('wilt');
INSERT INTO mindwell.invite_words ("word") VALUES('windlasses');
INSERT INTO mindwell.invite_words ("word") VALUES('yarely');
INSERT INTO mindwell.invite_words ("word") VALUES('yerked');
INSERT INTO mindwell.invite_words ("word") VALUES('yoeman');
INSERT INTO mindwell.invite_words ("word") VALUES('younker');

INSERT INTO mindwell.invite_words ("word") VALUES('aa');
INSERT INTO mindwell.invite_words ("word") VALUES('abaya');
INSERT INTO mindwell.invite_words ("word") VALUES('abomasum');
INSERT INTO mindwell.invite_words ("word") VALUES('absquatulate');
INSERT INTO mindwell.invite_words ("word") VALUES('adscititious');
INSERT INTO mindwell.invite_words ("word") VALUES('afreet');
INSERT INTO mindwell.invite_words ("word") VALUES('alcazar');
INSERT INTO mindwell.invite_words ("word") VALUES('amphibology');
INSERT INTO mindwell.invite_words ("word") VALUES('amphisbaena');
INSERT INTO mindwell.invite_words ("word") VALUES('anfractuous');
INSERT INTO mindwell.invite_words ("word") VALUES('anguilliform');
INSERT INTO mindwell.invite_words ("word") VALUES('apoptosis');
INSERT INTO mindwell.invite_words ("word") VALUES('argute');
INSERT INTO mindwell.invite_words ("word") VALUES('ariel');
INSERT INTO mindwell.invite_words ("word") VALUES('aristotle');
INSERT INTO mindwell.invite_words ("word") VALUES('aspergillum');
INSERT INTO mindwell.invite_words ("word") VALUES('astrobleme');
INSERT INTO mindwell.invite_words ("word") VALUES('autotomy');
INSERT INTO mindwell.invite_words ("word") VALUES('badmash');
INSERT INTO mindwell.invite_words ("word") VALUES('bandoline');
INSERT INTO mindwell.invite_words ("word") VALUES('bardolatry');
INSERT INTO mindwell.invite_words ("word") VALUES('bashment');
INSERT INTO mindwell.invite_words ("word") VALUES('bawbee');
INSERT INTO mindwell.invite_words ("word") VALUES('benthos');
INSERT INTO mindwell.invite_words ("word") VALUES('bergschrund');
INSERT INTO mindwell.invite_words ("word") VALUES('bezoar');
INSERT INTO mindwell.invite_words ("word") VALUES('bibliopole');
INSERT INTO mindwell.invite_words ("word") VALUES('bindlestiff');
INSERT INTO mindwell.invite_words ("word") VALUES('bingle');
INSERT INTO mindwell.invite_words ("word") VALUES('blatherskite');
INSERT INTO mindwell.invite_words ("word") VALUES('boffola');
INSERT INTO mindwell.invite_words ("word") VALUES('boilover');
INSERT INTO mindwell.invite_words ("word") VALUES('borborygmus');
INSERT INTO mindwell.invite_words ("word") VALUES('breatharian');
INSERT INTO mindwell.invite_words ("word") VALUES('bruxism');
INSERT INTO mindwell.invite_words ("word") VALUES('bumbo');
INSERT INTO mindwell.invite_words ("word") VALUES('burnsides');
INSERT INTO mindwell.invite_words ("word") VALUES('cacoethes');
INSERT INTO mindwell.invite_words ("word") VALUES('callipygian');
INSERT INTO mindwell.invite_words ("word") VALUES('callithumpian');
INSERT INTO mindwell.invite_words ("word") VALUES('camisado');
INSERT INTO mindwell.invite_words ("word") VALUES('canorous');
INSERT INTO mindwell.invite_words ("word") VALUES('cantillate');
INSERT INTO mindwell.invite_words ("word") VALUES('carphology');
INSERT INTO mindwell.invite_words ("word") VALUES('catoptromancy');
INSERT INTO mindwell.invite_words ("word") VALUES('cereology');
INSERT INTO mindwell.invite_words ("word") VALUES('cerulean');
INSERT INTO mindwell.invite_words ("word") VALUES('chad');
INSERT INTO mindwell.invite_words ("word") VALUES('chalkdown');
INSERT INTO mindwell.invite_words ("word") VALUES('chanticleer');
INSERT INTO mindwell.invite_words ("word") VALUES('chiliad');
INSERT INTO mindwell.invite_words ("word") VALUES('claggy');
INSERT INTO mindwell.invite_words ("word") VALUES('clepsydra');
INSERT INTO mindwell.invite_words ("word") VALUES('colporteur');
INSERT INTO mindwell.invite_words ("word") VALUES('comess');
INSERT INTO mindwell.invite_words ("word") VALUES('commensalism');
INSERT INTO mindwell.invite_words ("word") VALUES('comminatory');
INSERT INTO mindwell.invite_words ("word") VALUES('concinnity');
INSERT INTO mindwell.invite_words ("word") VALUES('congius');
INSERT INTO mindwell.invite_words ("word") VALUES('conniption');
INSERT INTO mindwell.invite_words ("word") VALUES('constellate');
INSERT INTO mindwell.invite_words ("word") VALUES('coprolalia');
INSERT INTO mindwell.invite_words ("word") VALUES('coriaceous');
INSERT INTO mindwell.invite_words ("word") VALUES('couthy');
INSERT INTO mindwell.invite_words ("word") VALUES('criticaster');
INSERT INTO mindwell.invite_words ("word") VALUES('crore');
INSERT INTO mindwell.invite_words ("word") VALUES('crottle');
INSERT INTO mindwell.invite_words ("word") VALUES('croze');
INSERT INTO mindwell.invite_words ("word") VALUES('cryptozoology');
INSERT INTO mindwell.invite_words ("word") VALUES('cudbear');
INSERT INTO mindwell.invite_words ("word") VALUES('cupreous');
INSERT INTO mindwell.invite_words ("word") VALUES('cyanic');
INSERT INTO mindwell.invite_words ("word") VALUES('cybersquatting');
INSERT INTO mindwell.invite_words ("word") VALUES('dariole');
INSERT INTO mindwell.invite_words ("word") VALUES('deasil');
INSERT INTO mindwell.invite_words ("word") VALUES('decubitus');
INSERT INTO mindwell.invite_words ("word") VALUES('deedy');
INSERT INTO mindwell.invite_words ("word") VALUES('defervescence');
INSERT INTO mindwell.invite_words ("word") VALUES('deglutition');
INSERT INTO mindwell.invite_words ("word") VALUES('degust');
INSERT INTO mindwell.invite_words ("word") VALUES('deipnosophist');
INSERT INTO mindwell.invite_words ("word") VALUES('deracinate');
INSERT INTO mindwell.invite_words ("word") VALUES('deterge');
INSERT INTO mindwell.invite_words ("word") VALUES('didi');
INSERT INTO mindwell.invite_words ("word") VALUES('digerati');
INSERT INTO mindwell.invite_words ("word") VALUES('dight');
INSERT INTO mindwell.invite_words ("word") VALUES('discobolus');
INSERT INTO mindwell.invite_words ("word") VALUES('disembogue');
INSERT INTO mindwell.invite_words ("word") VALUES('disenthral');
INSERT INTO mindwell.invite_words ("word") VALUES('divagate');
INSERT INTO mindwell.invite_words ("word") VALUES('divaricate');
INSERT INTO mindwell.invite_words ("word") VALUES('donkeyman');
INSERT INTO mindwell.invite_words ("word") VALUES('doryphore');
INSERT INTO mindwell.invite_words ("word") VALUES('dotish');
INSERT INTO mindwell.invite_words ("word") VALUES('douceur');
INSERT INTO mindwell.invite_words ("word") VALUES('draff');
INSERT INTO mindwell.invite_words ("word") VALUES('dragoman');
INSERT INTO mindwell.invite_words ("word") VALUES('dumbsize');
INSERT INTO mindwell.invite_words ("word") VALUES('dwaal');
INSERT INTO mindwell.invite_words ("word") VALUES('ecdysiast');
INSERT INTO mindwell.invite_words ("word") VALUES('edacious');
INSERT INTO mindwell.invite_words ("word") VALUES('effable');
INSERT INTO mindwell.invite_words ("word") VALUES('emacity');
INSERT INTO mindwell.invite_words ("word") VALUES('emmetropia');
INSERT INTO mindwell.invite_words ("word") VALUES('empasm');
INSERT INTO mindwell.invite_words ("word") VALUES('ensorcell');
INSERT INTO mindwell.invite_words ("word") VALUES('entomophagy');
INSERT INTO mindwell.invite_words ("word") VALUES('erf');
INSERT INTO mindwell.invite_words ("word") VALUES('ergometer');
INSERT INTO mindwell.invite_words ("word") VALUES('erubescent');
INSERT INTO mindwell.invite_words ("word") VALUES('etui');
INSERT INTO mindwell.invite_words ("word") VALUES('eucatastrophe');
INSERT INTO mindwell.invite_words ("word") VALUES('eurhythmic');
INSERT INTO mindwell.invite_words ("word") VALUES('eviternity');
INSERT INTO mindwell.invite_words ("word") VALUES('exequies');
INSERT INTO mindwell.invite_words ("word") VALUES('exsanguine');
INSERT INTO mindwell.invite_words ("word") VALUES('extramundane');
INSERT INTO mindwell.invite_words ("word") VALUES('eyewater');
INSERT INTO mindwell.invite_words ("word") VALUES('famulus');
INSERT INTO mindwell.invite_words ("word") VALUES('fankle');
INSERT INTO mindwell.invite_words ("word") VALUES('fipple');
INSERT INTO mindwell.invite_words ("word") VALUES('flatline');
INSERT INTO mindwell.invite_words ("word") VALUES('flews');
INSERT INTO mindwell.invite_words ("word") VALUES('floccinaucinihilipilification');
INSERT INTO mindwell.invite_words ("word") VALUES('flocculent');
INSERT INTO mindwell.invite_words ("word") VALUES('forehanded');
INSERT INTO mindwell.invite_words ("word") VALUES('frondeur');
INSERT INTO mindwell.invite_words ("word") VALUES('fugacious');
INSERT INTO mindwell.invite_words ("word") VALUES('funambulist');
INSERT INTO mindwell.invite_words ("word") VALUES('furuncle');
INSERT INTO mindwell.invite_words ("word") VALUES('fuscous');
INSERT INTO mindwell.invite_words ("word") VALUES('futhark');
INSERT INTO mindwell.invite_words ("word") VALUES('futz');
INSERT INTO mindwell.invite_words ("word") VALUES('gaberlunzie');
INSERT INTO mindwell.invite_words ("word") VALUES('gaita');
INSERT INTO mindwell.invite_words ("word") VALUES('galligaskins');
INSERT INTO mindwell.invite_words ("word") VALUES('gallus');
INSERT INTO mindwell.invite_words ("word") VALUES('gasconade');
INSERT INTO mindwell.invite_words ("word") VALUES('glabrous');
INSERT INTO mindwell.invite_words ("word") VALUES('glaikit');
INSERT INTO mindwell.invite_words ("word") VALUES('gnathic');
INSERT INTO mindwell.invite_words ("word") VALUES('gobemouche');
INSERT INTO mindwell.invite_words ("word") VALUES('goodfella');
INSERT INTO mindwell.invite_words ("word") VALUES('guddle');
INSERT INTO mindwell.invite_words ("word") VALUES('habile');
INSERT INTO mindwell.invite_words ("word") VALUES('hallux');
INSERT INTO mindwell.invite_words ("word") VALUES('haruspex');
INSERT INTO mindwell.invite_words ("word") VALUES('higgler');
INSERT INTO mindwell.invite_words ("word") VALUES('hinky');
INSERT INTO mindwell.invite_words ("word") VALUES('hodiernal');
INSERT INTO mindwell.invite_words ("word") VALUES('hoggin');
INSERT INTO mindwell.invite_words ("word") VALUES('hongi');
INSERT INTO mindwell.invite_words ("word") VALUES('howff');
INSERT INTO mindwell.invite_words ("word") VALUES('humdudgeon');
INSERT INTO mindwell.invite_words ("word") VALUES('hwyl');
INSERT INTO mindwell.invite_words ("word") VALUES('illywhacker');
INSERT INTO mindwell.invite_words ("word") VALUES('incrassate');
INSERT INTO mindwell.invite_words ("word") VALUES('incunabula');
INSERT INTO mindwell.invite_words ("word") VALUES('ingurgitate');
INSERT INTO mindwell.invite_words ("word") VALUES('inspissate');
INSERT INTO mindwell.invite_words ("word") VALUES('inunct');
INSERT INTO mindwell.invite_words ("word") VALUES('jumbuck');
INSERT INTO mindwell.invite_words ("word") VALUES('jumentous');
INSERT INTO mindwell.invite_words ("word") VALUES('jungli');
INSERT INTO mindwell.invite_words ("word") VALUES('karateka');
INSERT INTO mindwell.invite_words ("word") VALUES('keek');
INSERT INTO mindwell.invite_words ("word") VALUES('kenspeckle');
INSERT INTO mindwell.invite_words ("word") VALUES('kinnikinnick');
INSERT INTO mindwell.invite_words ("word") VALUES('kylie');
INSERT INTO mindwell.invite_words ("word") VALUES('labarum');
INSERT INTO mindwell.invite_words ("word") VALUES('lablab');
INSERT INTO mindwell.invite_words ("word") VALUES('lactarium');
INSERT INTO mindwell.invite_words ("word") VALUES('liripipe');
INSERT INTO mindwell.invite_words ("word") VALUES('loblolly');
INSERT INTO mindwell.invite_words ("word") VALUES('lobola');
INSERT INTO mindwell.invite_words ("word") VALUES('logomachy');
INSERT INTO mindwell.invite_words ("word") VALUES('lollygag');
INSERT INTO mindwell.invite_words ("word") VALUES('luculent');
INSERT INTO mindwell.invite_words ("word") VALUES('lycanthropy');
INSERT INTO mindwell.invite_words ("word") VALUES('macushla');
INSERT INTO mindwell.invite_words ("word") VALUES('mallam');
INSERT INTO mindwell.invite_words ("word") VALUES('mamaguy');
INSERT INTO mindwell.invite_words ("word") VALUES('martlet');
INSERT INTO mindwell.invite_words ("word") VALUES('meacock');
INSERT INTO mindwell.invite_words ("word") VALUES('merkin');
INSERT INTO mindwell.invite_words ("word") VALUES('merrythought');
INSERT INTO mindwell.invite_words ("word") VALUES('mim');
INSERT INTO mindwell.invite_words ("word") VALUES('mimsy');
INSERT INTO mindwell.invite_words ("word") VALUES('minacious');
INSERT INTO mindwell.invite_words ("word") VALUES('minibeast');
INSERT INTO mindwell.invite_words ("word") VALUES('misogamy');
INSERT INTO mindwell.invite_words ("word") VALUES('mistigris');
INSERT INTO mindwell.invite_words ("word") VALUES('mixologist');
INSERT INTO mindwell.invite_words ("word") VALUES('mollitious');
INSERT INTO mindwell.invite_words ("word") VALUES('momism');
INSERT INTO mindwell.invite_words ("word") VALUES('monorchid');
INSERT INTO mindwell.invite_words ("word") VALUES('moonraker');
INSERT INTO mindwell.invite_words ("word") VALUES('mudlark');
INSERT INTO mindwell.invite_words ("word") VALUES('muktuk');
INSERT INTO mindwell.invite_words ("word") VALUES('mumpsimus');
INSERT INTO mindwell.invite_words ("word") VALUES('nacarat');
INSERT INTO mindwell.invite_words ("word") VALUES('nagware');
INSERT INTO mindwell.invite_words ("word") VALUES('nainsook');
INSERT INTO mindwell.invite_words ("word") VALUES('natation');
INSERT INTO mindwell.invite_words ("word") VALUES('nesh');
INSERT INTO mindwell.invite_words ("word") VALUES('netizen');
INSERT INTO mindwell.invite_words ("word") VALUES('noctambulist');
INSERT INTO mindwell.invite_words ("word") VALUES('noyade');
INSERT INTO mindwell.invite_words ("word") VALUES('nugacity');
INSERT INTO mindwell.invite_words ("word") VALUES('nympholepsy');
INSERT INTO mindwell.invite_words ("word") VALUES('obnubilate');
INSERT INTO mindwell.invite_words ("word") VALUES('ogdoad');
INSERT INTO mindwell.invite_words ("word") VALUES('omophagy');
INSERT INTO mindwell.invite_words ("word") VALUES('omphalos');
INSERT INTO mindwell.invite_words ("word") VALUES('onolatry');
INSERT INTO mindwell.invite_words ("word") VALUES('operose');
INSERT INTO mindwell.invite_words ("word") VALUES('opsimath');
INSERT INTO mindwell.invite_words ("word") VALUES('orectic');
INSERT INTO mindwell.invite_words ("word") VALUES('orrery');
INSERT INTO mindwell.invite_words ("word") VALUES('ortanique');
INSERT INTO mindwell.invite_words ("word") VALUES('otalgia');
INSERT INTO mindwell.invite_words ("word") VALUES('oxter');
INSERT INTO mindwell.invite_words ("word") VALUES('paludal');
INSERT INTO mindwell.invite_words ("word") VALUES('panurgic');
INSERT INTO mindwell.invite_words ("word") VALUES('parapente');
INSERT INTO mindwell.invite_words ("word") VALUES('paraph');
INSERT INTO mindwell.invite_words ("word") VALUES('patulous');
INSERT INTO mindwell.invite_words ("word") VALUES('pavonine');
INSERT INTO mindwell.invite_words ("word") VALUES('pedicular');
INSERT INTO mindwell.invite_words ("word") VALUES('peever');
INSERT INTO mindwell.invite_words ("word") VALUES('periapt');
INSERT INTO mindwell.invite_words ("word") VALUES('petcock');
INSERT INTO mindwell.invite_words ("word") VALUES('peterman');
INSERT INTO mindwell.invite_words ("word") VALUES('pettitoes');
INSERT INTO mindwell.invite_words ("word") VALUES('piacular');
INSERT INTO mindwell.invite_words ("word") VALUES('pilgarlic');
INSERT INTO mindwell.invite_words ("word") VALUES('pinguid');
INSERT INTO mindwell.invite_words ("word") VALUES('piscatorial');
INSERT INTO mindwell.invite_words ("word") VALUES('pleurodynia');
INSERT INTO mindwell.invite_words ("word") VALUES('plew');
INSERT INTO mindwell.invite_words ("word") VALUES('pogey');
INSERT INTO mindwell.invite_words ("word") VALUES('pollex');
INSERT INTO mindwell.invite_words ("word") VALUES('pooter');
INSERT INTO mindwell.invite_words ("word") VALUES('portolan');
INSERT INTO mindwell.invite_words ("word") VALUES('posology');
INSERT INTO mindwell.invite_words ("word") VALUES('possident');
INSERT INTO mindwell.invite_words ("word") VALUES('pother');
INSERT INTO mindwell.invite_words ("word") VALUES('presenteeism');
INSERT INTO mindwell.invite_words ("word") VALUES('previse');
INSERT INTO mindwell.invite_words ("word") VALUES('probang');
INSERT INTO mindwell.invite_words ("word") VALUES('prosopagnosia');
INSERT INTO mindwell.invite_words ("word") VALUES('puddysticks');
INSERT INTO mindwell.invite_words ("word") VALUES('pyknic');
INSERT INTO mindwell.invite_words ("word") VALUES('pyroclastic');
INSERT INTO mindwell.invite_words ("word") VALUES('ragtop');
INSERT INTO mindwell.invite_words ("word") VALUES('ratite');
INSERT INTO mindwell.invite_words ("word") VALUES('rawky');
INSERT INTO mindwell.invite_words ("word") VALUES('razzia');
INSERT INTO mindwell.invite_words ("word") VALUES('rebirthing');
INSERT INTO mindwell.invite_words ("word") VALUES('retiform');
INSERT INTO mindwell.invite_words ("word") VALUES('rhinoplasty');
INSERT INTO mindwell.invite_words ("word") VALUES('rubiginous');
INSERT INTO mindwell.invite_words ("word") VALUES('rubricate');
INSERT INTO mindwell.invite_words ("word") VALUES('rumpot');
INSERT INTO mindwell.invite_words ("word") VALUES('sangoma');
INSERT INTO mindwell.invite_words ("word") VALUES('sarmie');
INSERT INTO mindwell.invite_words ("word") VALUES('saucier');
INSERT INTO mindwell.invite_words ("word") VALUES('saudade');
INSERT INTO mindwell.invite_words ("word") VALUES('scofflaw');
INSERT INTO mindwell.invite_words ("word") VALUES('screenager');
INSERT INTO mindwell.invite_words ("word") VALUES('scrippage');
INSERT INTO mindwell.invite_words ("word") VALUES('selkie');
INSERT INTO mindwell.invite_words ("word") VALUES('serac');
INSERT INTO mindwell.invite_words ("word") VALUES('sesquipedalian');
INSERT INTO mindwell.invite_words ("word") VALUES('shallop');
INSERT INTO mindwell.invite_words ("word") VALUES('shamal');
INSERT INTO mindwell.invite_words ("word") VALUES('shavetail');
INSERT INTO mindwell.invite_words ("word") VALUES('shippon');
INSERT INTO mindwell.invite_words ("word") VALUES('shofar');
INSERT INTO mindwell.invite_words ("word") VALUES('skanky');
INSERT INTO mindwell.invite_words ("word") VALUES('skelf');
INSERT INTO mindwell.invite_words ("word") VALUES('skimmington');
INSERT INTO mindwell.invite_words ("word") VALUES('skycap');
INSERT INTO mindwell.invite_words ("word") VALUES('snakebitten');
INSERT INTO mindwell.invite_words ("word") VALUES('snollygoster');
INSERT INTO mindwell.invite_words ("word") VALUES('sockdolager');
INSERT INTO mindwell.invite_words ("word") VALUES('solander');
INSERT INTO mindwell.invite_words ("word") VALUES('soucouyant');
INSERT INTO mindwell.invite_words ("word") VALUES('spaghettification');
INSERT INTO mindwell.invite_words ("word") VALUES('spitchcock');
INSERT INTO mindwell.invite_words ("word") VALUES('splanchnic');
INSERT INTO mindwell.invite_words ("word") VALUES('spurrier');
INSERT INTO mindwell.invite_words ("word") VALUES('stercoraceous');
INSERT INTO mindwell.invite_words ("word") VALUES('sternutator');
INSERT INTO mindwell.invite_words ("word") VALUES('stiction');
INSERT INTO mindwell.invite_words ("word") VALUES('strappado');
INSERT INTO mindwell.invite_words ("word") VALUES('strigil');
INSERT INTO mindwell.invite_words ("word") VALUES('struthious');
INSERT INTO mindwell.invite_words ("word") VALUES('studmuffin');
INSERT INTO mindwell.invite_words ("word") VALUES('stylite');
INSERT INTO mindwell.invite_words ("word") VALUES('subfusc');
INSERT INTO mindwell.invite_words ("word") VALUES('submontane');
INSERT INTO mindwell.invite_words ("word") VALUES('succuss');
INSERT INTO mindwell.invite_words ("word") VALUES('sudd');
INSERT INTO mindwell.invite_words ("word") VALUES('suedehead');
INSERT INTO mindwell.invite_words ("word") VALUES('superbious');
INSERT INTO mindwell.invite_words ("word") VALUES('superette');
INSERT INTO mindwell.invite_words ("word") VALUES('taniwha');
INSERT INTO mindwell.invite_words ("word") VALUES('tappen');
INSERT INTO mindwell.invite_words ("word") VALUES('tellurian');
INSERT INTO mindwell.invite_words ("word") VALUES('testudo');
INSERT INTO mindwell.invite_words ("word") VALUES('thalassic');
INSERT INTO mindwell.invite_words ("word") VALUES('thaumatrope');
INSERT INTO mindwell.invite_words ("word") VALUES('thirstland');
INSERT INTO mindwell.invite_words ("word") VALUES('thrutch');
INSERT INTO mindwell.invite_words ("word") VALUES('thurifer');
INSERT INTO mindwell.invite_words ("word") VALUES('tiffin');
INSERT INTO mindwell.invite_words ("word") VALUES('tigon');
INSERT INTO mindwell.invite_words ("word") VALUES('tokoloshe');
INSERT INTO mindwell.invite_words ("word") VALUES('toplofty');
INSERT INTO mindwell.invite_words ("word") VALUES('transpicuous');
INSERT INTO mindwell.invite_words ("word") VALUES('triskaidekaphobia');
INSERT INTO mindwell.invite_words ("word") VALUES('triskelion');
INSERT INTO mindwell.invite_words ("word") VALUES('tsantsa');
INSERT INTO mindwell.invite_words ("word") VALUES('turbary');
INSERT INTO mindwell.invite_words ("word") VALUES('ulu');
INSERT INTO mindwell.invite_words ("word") VALUES('umbriferous');
INSERT INTO mindwell.invite_words ("word") VALUES('uncinate');
INSERT INTO mindwell.invite_words ("word") VALUES('uniped');
INSERT INTO mindwell.invite_words ("word") VALUES('uroboros');
INSERT INTO mindwell.invite_words ("word") VALUES('ustad');
INSERT INTO mindwell.invite_words ("word") VALUES('vagarious');
INSERT INTO mindwell.invite_words ("word") VALUES('velleity');
INSERT INTO mindwell.invite_words ("word") VALUES('verjuice');
INSERT INTO mindwell.invite_words ("word") VALUES('vicinal');
INSERT INTO mindwell.invite_words ("word") VALUES('vidiot');
INSERT INTO mindwell.invite_words ("word") VALUES('vomitous');
INSERT INTO mindwell.invite_words ("word") VALUES('wabbit');
INSERT INTO mindwell.invite_words ("word") VALUES('waitron');
INSERT INTO mindwell.invite_words ("word") VALUES('wakeboarding');
INSERT INTO mindwell.invite_words ("word") VALUES('wayzgoose');
INSERT INTO mindwell.invite_words ("word") VALUES('winebibber');
INSERT INTO mindwell.invite_words ("word") VALUES('wittol');
INSERT INTO mindwell.invite_words ("word") VALUES('woopie');
INSERT INTO mindwell.invite_words ("word") VALUES('wowser');
INSERT INTO mindwell.invite_words ("word") VALUES('xenology');
INSERT INTO mindwell.invite_words ("word") VALUES('ylem');
INSERT INTO mindwell.invite_words ("word") VALUES('zetetic');
INSERT INTO mindwell.invite_words ("word") VALUES('zoolatry');
INSERT INTO mindwell.invite_words ("word") VALUES('zopissa');
INSERT INTO mindwell.invite_words ("word") VALUES('zorro');



-- CREATE TABLE "invites" ------------------------------------
CREATE TABLE "mindwell"."invites" (
    "id" Serial NOT NULL,
    "referrer_id" Integer NOT NULL,
    "word1" Integer NOT NULL,
    "word2" Integer NOT NULL,
    "word3" Integer NOT NULL,
	"created_at" Date DEFAULT CURRENT_DATE NOT NULL,
    CONSTRAINT "unique_invite_id" PRIMARY KEY( "id" ),
    CONSTRAINT "invite_word1" FOREIGN KEY("word1") REFERENCES "mindwell"."invite_words"("id"),
    CONSTRAINT "invite_word2" FOREIGN KEY("word2") REFERENCES "mindwell"."invite_words"("id"),
    CONSTRAINT "invite_word3" FOREIGN KEY("word3") REFERENCES "mindwell"."invite_words"("id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_referrer_id" ---------------------------
CREATE INDEX "index_referrer_id" ON "mindwell"."invites" USING btree( "referrer_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_invite_words" ---------------------------
CREATE UNIQUE INDEX "index_invite_words" ON "mindwell"."invites" USING btree( "word1", "word2", "word3" );
-- -------------------------------------------------------------

INSERT INTO mindwell.invites (referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1);
INSERT INTO mindwell.invites (referrer_id, word1, word2, word3) VALUES(1, 2, 2, 2);
INSERT INTO mindwell.invites (referrer_id, word1, word2, word3) VALUES(1, 3, 3, 3);

CREATE OR REPLACE FUNCTION give_invite(userName TEXT) RETURNS VOID AS $$
    DECLARE
        wordCount INTEGER;
        userId INTEGER;
    BEGIN
        wordCount = (SELECT COUNT(*) FROM invite_words);
        userId = (SELECT id FROM users WHERE lower(name) = lower(userName));

        INSERT INTO invites(referrer_id, word1, word2, word3)
            VALUES(userId, 
                trunc(random() * wordCount),
                trunc(random() * wordCount),
                trunc(random() * wordCount));
    END;
$$ LANGUAGE plpgsql;

CREATE VIEW mindwell.unwrapped_invites AS
SELECT invites.id AS id, 
    users.id AS user_id,
    lower(users.name) AS name, 
    one.word AS word1, 
    two.word AS word2, 
    three.word AS word3
FROM mindwell.invites, mindwell.users,
    mindwell.invite_words AS one,
    mindwell.invite_words AS two,
    mindwell.invite_words AS three
WHERE invites.referrer_id = users.id
    AND invites.word1 = one.id 
    AND invites.word2 = two.id 
    AND invites.word3 = three.id;



-- CREATE TABLE "relation" -------------------------------------
CREATE TABLE "mindwell"."relation" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."relation" VALUES(0, 'followed');
INSERT INTO "mindwell"."relation" VALUES(1, 'requested');
INSERT INTO "mindwell"."relation" VALUES(2, 'cancelled');
INSERT INTO "mindwell"."relation" VALUES(3, 'ignored');
INSERT INTO "mindwell"."relation" VALUES(4, 'hidden');
-- -------------------------------------------------------------



-- CREATE TABLE "relations" ------------------------------------
CREATE TABLE "mindwell"."relations" (
	"from_id" Integer NOT NULL,
	"to_id" Integer NOT NULL,
	"type" Integer NOT NULL,
	"changed_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT "unique_relation" PRIMARY KEY ("from_id" , "to_id"),
    CONSTRAINT "unique_from_relation" FOREIGN KEY ("from_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "unique_to_relation" FOREIGN KEY ("to_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "enum_relation_type" FOREIGN KEY("type") REFERENCES "mindwell"."relation"("id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_related_to_users" -----------------------
CREATE INDEX "index_related_to_users" ON "mindwell"."relations" USING btree( "to_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_related_from_users" ---------------------
CREATE INDEX "index_related_from_users" ON "mindwell"."relations" USING btree( "from_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.count_relations_ins() RETURNS TRIGGER AS $$
    BEGIN
        IF (NEW."type" = (SELECT id FROM mindwell.relation WHERE "type" = 'followed')) THEN
            UPDATE mindwell.users
            SET followers_count = followers_count + 1
            WHERE id = NEW.to_id;
            UPDATE mindwell.users
            SET followings_count = followings_count + 1
            WHERE id = NEW.from_id;
        ELSIF (NEW."type" = (SELECT id FROM mindwell.relation WHERE "type" = 'ignored')) THEN
            UPDATE mindwell.users
            SET ignored_count = ignored_count + 1
            WHERE id = NEW.from_id;
        END IF;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_relations_ins
    AFTER INSERT OR UPDATE ON mindwell.relations
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_relations_ins();

CREATE OR REPLACE FUNCTION mindwell.count_relations_del() RETURNS TRIGGER AS $$
    BEGIN
        IF (OLD."type" = (SELECT id FROM mindwell.relation WHERE "type" = 'followed')) THEN
            UPDATE mindwell.users
            SET followers_count = followers_count - 1
            WHERE id = OLD.to_id;
            UPDATE mindwell.users
            SET followings_count = followings_count - 1
            WHERE id = OLD.from_id;
        ELSIF (OLD."type" = (SELECT id FROM mindwell.relation WHERE "type" = 'ignored')) THEN
            UPDATE users
            SET ignored_count = ignored_count - 1
            WHERE id = OLD.from_id;
        END IF;
    
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_relations_del
    AFTER UPDATE OR DELETE ON mindwell.relations
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_relations_del();

CREATE OR REPLACE FUNCTION mindwell.del_relation_from_ignored() RETURNS TRIGGER AS $$
    DECLARE
        ignored Integer;
    BEGIN
        ignored = (SELECT id FROM mindwell.relation WHERE "type" = 'ignored');

        IF (NEW."type" = ignored) THEN
            DELETE FROM relations
            WHERE relations.from_id = NEW.to_id 
                AND relations.to_id = NEW.from_id
                AND relations."type" != ignored;
        END IF;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER relation_from_ignored
    AFTER INSERT OR UPDATE ON mindwell.relations
    FOR EACH ROW EXECUTE PROCEDURE mindwell.del_relation_from_ignored();



-- CREATE TABLE "entry_privacy" --------------------------------
CREATE TABLE "mindwell"."entry_privacy" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."entry_privacy" VALUES(0, 'all');
INSERT INTO "mindwell"."entry_privacy" VALUES(1, 'some');
INSERT INTO "mindwell"."entry_privacy" VALUES(2, 'me');
INSERT INTO "mindwell"."entry_privacy" VALUES(3, 'anonymous');
-- -------------------------------------------------------------



-- CREATE TABLE "categories" -----------------------------------
CREATE TABLE "mindwell"."categories" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."categories" VALUES(0, 'tweet');
INSERT INTO "mindwell"."categories" VALUES(1, 'longread');
INSERT INTO "mindwell"."categories" VALUES(2, 'media');
INSERT INTO "mindwell"."categories" VALUES(3, 'comment');
-- -------------------------------------------------------------



-- CREATE TABLE "entries" --------------------------------------
CREATE TABLE "mindwell"."entries" (
	"id" Serial NOT NULL,
	"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	"author_id" Integer NOT NULL,
	"title" Text DEFAULT '' NOT NULL,
    "cut_title" TEXT DEFAULT '' NOT NULL,
	"content" Text NOT NULL,
    "cut_content" TEXT DEFAULT '' NOT NULL,
    "edit_content" Text NOT NULL,
    "has_cut" BOOLEAN DEFAULT FALSE NOT NULL,
	"word_count" Integer NOT NULL,
	"visible_for" Integer NOT NULL,
	"is_votable" Boolean NOT NULL,
    "in_live" Boolean DEFAULT TRUE NOT NULL,
	"rating" Real DEFAULT 0 NOT NULL,
    "up_votes" Integer DEFAULT 0 NOT NULL,
    "down_votes" Integer DEFAULT 0 NOT NULL,
    "vote_sum" Real DEFAULT 0 NOT NULL,
    "weight_sum" Real DEFAULT 0 NOT NULL,
    "category" Integer NOT NULL,
	"comments_count" Integer DEFAULT 0 NOT NULL,
    "last_comment" Integer,
	CONSTRAINT "unique_entry_id" PRIMARY KEY( "id" ),
    CONSTRAINT "entry_user_id" FOREIGN KEY("author_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "enum_entry_privacy" FOREIGN KEY("visible_for") REFERENCES "mindwell"."entry_privacy"("id"),
    CONSTRAINT "entry_category" FOREIGN KEY("category") REFERENCES "mindwell"."categories"("id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_id" -------------------------------
CREATE INDEX "index_entry_id" ON "mindwell"."entries" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_date" -----------------------------
CREATE INDEX "index_entry_date" ON "mindwell"."entries" USING btree( "created_at" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_users_id" -------------------------
CREATE INDEX "index_entry_users_id" ON "mindwell"."entries" USING btree( "author_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_rating" ---------------------------
CREATE INDEX "index_entry_rating" ON "mindwell"."entries" USING btree( "rating" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_word_count" -----------------------
CREATE INDEX "index_entry_word_count" ON "mindwell"."entries" USING btree( "word_count" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_last_comment_id" ------------------------
CREATE INDEX "index_last_comment_id" ON "mindwell"."entries" USING btree( "last_comment" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.inc_tlog_entries() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET entries_count = entries_count + 1
        WHERE id = NEW.author_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.dec_tlog_entries() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET entries_count = entries_count - 1
        WHERE id = OLD.author_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_tlog_entries_ins
    AFTER INSERT ON mindwell.entries
    FOR EACH ROW 
    WHEN (NEW.visible_for = 0) -- visible_for = all
    EXECUTE PROCEDURE mindwell.inc_tlog_entries();

CREATE TRIGGER cnt_tlog_entries_upd_inc
    AFTER UPDATE ON mindwell.entries
    FOR EACH ROW 
    WHEN (OLD.visible_for <> 0 AND NEW.visible_for = 0)
    EXECUTE PROCEDURE mindwell.inc_tlog_entries();

CREATE TRIGGER cnt_tlog_entries_upd_dec
    AFTER UPDATE ON mindwell.entries
    FOR EACH ROW 
    WHEN (OLD.visible_for = 0 AND NEW.visible_for <> 0)
    EXECUTE PROCEDURE mindwell.dec_tlog_entries();

CREATE TRIGGER cnt_tlog_entries_del
    AFTER DELETE ON mindwell.entries
    FOR EACH ROW 
    WHEN (OLD.visible_for = 0)
    EXECUTE PROCEDURE mindwell.dec_tlog_entries();



-- CREATE TABLE "tags" -----------------------------------------
CREATE TABLE "mindwell"."tags" (
    "id" Serial NOT NULL,
    "tag" Text NOT NULL,
    CONSTRAINT "unique_tag_id" PRIMARY KEY( "id" ) );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_tag" ------------------------------------
CREATE UNIQUE INDEX "index_tag" ON "mindwell"."tags" USING btree( "tag" ) ;
-- -------------------------------------------------------------



-- CREATE TABLE "entry_tags" -----------------------------------
CREATE TABLE "mindwell"."entry_tags" (
    "entry_id" Integer NOT NULL,
    "tag_id" Integer NOT NULL,
    CONSTRAINT "entry_tags_entry" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "entry_tags_tag" FOREIGN KEY("tag_id") REFERENCES "mindwell"."tags"("id"),
    CONSTRAINT "unique_entry_tag" UNIQUE("entry_id", "tag_id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_tag" ------------------------------------
CREATE INDEX "index_entry_tags_entry" ON "mindwell"."entry_tags" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_tag" ------------------------------------
CREATE INDEX "index_entry_tags_tag" ON "mindwell"."entry_tags" USING btree( "tag_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.count_tags() RETURNS TRIGGER AS $$
    BEGIN
        WITH authors AS 
        (
            SELECT DISTINCT author_id as id
            FROM mindwell.entries, changes
            WHERE entries.id = changes.entry_id
        )
        UPDATE mindwell.users
        SET tags_count = counts.cnt 
        FROM authors,
        (
            SELECT author_id, COUNT(tag_id) as cnt
            FROM mindwell.entries, mindwell.entry_tags, authors
            WHERE authors.id = entries.author_id AND entries.id = entry_tags.entry_id
            GROUP BY author_id
        ) AS counts
        WHERE authors.id = users.id AND counts.author_id = users.id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_tags_ins
    AFTER INSERT ON mindwell.entry_tags
    REFERENCING NEW TABLE as changes
    FOR EACH STATEMENT EXECUTE PROCEDURE mindwell.count_tags();

CREATE TRIGGER cnt_tags_del
    AFTER DELETE ON mindwell.entry_tags
    REFERENCING OLD TABLE as changes
    FOR EACH STATEMENT EXECUTE PROCEDURE mindwell.count_tags();



-- CREATE TABLE "favorites" ------------------------------------
CREATE TABLE "mindwell"."favorites" (
	"user_id" Integer NOT NULL,
	"entry_id" Integer NOT NULL,
    "date" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "favorite_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "favorite_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_user_favorite" UNIQUE("user_id", "entry_id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_favorite_entries" -----------------------
CREATE INDEX "index_favorite_entries" ON "mindwell"."favorites" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_favorite_users" -------------------------
CREATE INDEX "index_favorite_users" ON "mindwell"."favorites" USING btree( "user_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.inc_favorites() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET favorites_count = favorites_count + 1
        WHERE id = NEW.user_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.dec_favorites() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET favorites_count = favorites_count - 1
        WHERE id = OLD.user_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_favorites_inc
    AFTER INSERT ON mindwell.favorites
    FOR EACH ROW EXECUTE PROCEDURE mindwell.inc_favorites();

CREATE TRIGGER cnt_favorites_dec
    AFTER DELETE ON mindwell.favorites
    FOR EACH ROW EXECUTE PROCEDURE mindwell.dec_favorites();



-- CREATE TABLE "watching" -------------------------------------
CREATE TABLE "mindwell"."watching" (
	"user_id" Integer NOT NULL,
	"entry_id" Integer NOT NULL,
    "date" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "watching_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "watching_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_user_watching" UNIQUE("user_id", "entry_id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_watching_entries" -----------------------
CREATE INDEX "index_watching_entries" ON "mindwell"."watching" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_watching_users" -------------------------
CREATE INDEX "index_watching_users" ON "mindwell"."watching" USING btree( "user_id" );
-- -------------------------------------------------------------



-- CREATE TABLE "entry_votes" ----------------------------------
CREATE TABLE "mindwell"."entry_votes" (
	"user_id" Integer NOT NULL,
	"entry_id" Integer NOT NULL,
    "vote" Real NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "entry_vote_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "entry_vote_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_entry_vote" UNIQUE("user_id", "entry_id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_voted_entries" --------------------------
CREATE INDEX "index_voted_entries" ON "mindwell"."entry_votes" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_voted_users" ----------------------------
CREATE INDEX "index_voted_users" ON "mindwell"."entry_votes" USING btree( "user_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.entry_votes_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET up_votes = up_votes + (NEW.vote > 0)::int,
            down_votes = down_votes + (NEW.vote < 0)::int,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            rating = atan2(weight_sum + abs(NEW.vote), 2)
                * (vote_sum + NEW.vote) / (weight_sum + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count + 1,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            weight = atan2(vote_count + 1, 20) * (vote_sum + NEW.vote) 
                / (weight_sum + abs(NEW.vote)) / pi() * 2
        FROM entry
        WHERE user_id = entry.author_id 
            AND vote_weights.category = entry.category;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET up_votes = up_votes - (OLD.vote > 0)::int + (NEW.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int + (NEW.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            rating = atan2(weight_sum - abs(OLD.vote) + abs(NEW.vote), 2)
                * (vote_sum - OLD.vote + NEW.vote) / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = NEW.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            weight = atan2(vote_count, 20) * (vote_sum - OLD.vote + NEW.vote) 
                / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 2
        FROM entry
        WHERE user_id = entry.author_id
            AND vote_weights.category = entry.category;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.entry_votes_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET up_votes = up_votes - (OLD.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            rating = CASE WHEN weight_sum = abs(OLD.vote) THEN 0
                ELSE atan2(weight_sum - abs(OLD.vote), 2)
                    * (vote_sum - OLD.vote) / (weight_sum - abs(OLD.vote)) / pi() * 200
                END
        WHERE id = OLD.entry_id;
        
        WITH entry AS (
            SELECT author_id, category
            FROM mindwell.entries
            WHERE id = OLD.entry_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count - 1,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            weight = CASE WHEN weight_sum = abs(OLD.vote) THEN 0.1
                ELSE atan2(vote_count - 1, 20) * (vote_sum - OLD.vote) 
                    / (weight_sum - abs(OLD.vote)) / pi() * 2
                END
        FROM entry
        WHERE user_id = entry.author_id
            AND vote_weights.category = entry.category;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_entry_votes_ins
    AFTER INSERT ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.entry_votes_ins();

CREATE TRIGGER cnt_entry_votes_upd
    AFTER UPDATE ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.entry_votes_upd();

CREATE TRIGGER cnt_entry_votes_del
    AFTER DELETE ON mindwell.entry_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.entry_votes_del();



-- CREATE TABLE "vote_weights" ---------------------------------
CREATE TABLE "mindwell"."vote_weights" (
	"user_id" Integer NOT NULL,
	"category" Integer NOT NULL,
    "weight" Real DEFAULT 0.1 NOT NULL,
    "vote_count" Integer DEFAULT 0 NOT NULL,
    "vote_sum" Real DEFAULT 0 NOT NULL,
    "weight_sum" Real DEFAULT 0 NOT NULL,
    CONSTRAINT "vote_weights_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "vote_weights_category" FOREIGN KEY("category") REFERENCES "mindwell"."categories"("id"),
    CONSTRAINT "unique_vote_weight" UNIQUE("user_id", "category") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_vote_weights" ---------------------------
CREATE INDEX "index_vote_weights" ON "mindwell"."vote_weights" USING btree( "user_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.create_vote_weights() RETURNS TRIGGER AS $$
    BEGIN
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 0);
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 1);
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 2);
        INSERT INTO mindwell.vote_weights(user_id, category) VALUES(NEW.id, 3);

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER crt_vote_weights
    AFTER INSERT ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.create_vote_weights();



-- CREATE TABLE "entries_privacy" ------------------------------
CREATE TABLE "mindwell"."entries_privacy" (
	"user_id" Integer NOT NULL,
	"entry_id" Integer NOT NULL,
    CONSTRAINT "entries_privacy_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "entries_privacy_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_entry_privacy" UNIQUE("user_id", "entry_id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_private_entries" ------------------------
CREATE INDEX "index_private_entries" ON "mindwell"."entries_privacy" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_private_users" --------------------------
CREATE INDEX "index_private_users" ON "mindwell"."entries_privacy" USING btree( "user_id" );
-- -------------------------------------------------------------



-- CREATE TABLE "comments" -------------------------------------
CREATE TABLE "mindwell"."comments" (
	"id" Serial NOT NULL,
	"author_id" Integer NOT NULL,
	"entry_id" Integer NOT NULL,
	"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
	"content" Text NOT NULL,
    "edit_content" Text DEFAULT '' NOT NULL,
	"rating" Real DEFAULT 0 NOT NULL,
    "up_votes" Integer DEFAULT 0 NOT NULL,
    "down_votes" Integer DEFAULT 0 NOT NULL,
    "vote_sum" Real DEFAULT 0 NOT NULL,
    "weight_sum" Real DEFAULT 0 NOT NULL,
	CONSTRAINT "unique_comment_id" PRIMARY KEY( "id" ),
    CONSTRAINT "comment_user_id" FOREIGN KEY("author_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "comment_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_comment_entry_id" -----------------------
CREATE INDEX "index_comment_entry_id" ON "mindwell"."comments" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_comment_date" ---------------------------
CREATE INDEX "index_comment_date" ON "mindwell"."comments" USING btree( "created_at" );
-- -------------------------------------------------------------

ALTER TABLE "mindwell"."entries"
ADD CONSTRAINT "entry_last_comment_id" FOREIGN KEY("last_comment") REFERENCES "mindwell"."comments"("id");

CREATE OR REPLACE FUNCTION mindwell.inc_comments() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET comments_count = comments_count + 1
        WHERE id = NEW.author_id;
        
        UPDATE mindwell.entries
        SET comments_count = comments_count + 1
        WHERE id = NEW.entry_id;
        
        INSERT INTO mindwell.watching
        VALUES(NEW.author_id, NEW.entry_id)
        ON CONFLICT ON CONSTRAINT unique_user_watching DO NOTHING;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_comments_inc
    AFTER INSERT ON mindwell.comments
    FOR EACH ROW EXECUTE PROCEDURE mindwell.inc_comments();

CREATE OR REPLACE FUNCTION mindwell.dec_comments() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET comments_count = comments_count - 1
        WHERE id = OLD.author_id;
        
        UPDATE mindwell.entries
        SET comments_count = comments_count - 1
        WHERE id = OLD.entry_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_comments_dec
    AFTER DELETE ON mindwell.comments
    FOR EACH ROW EXECUTE PROCEDURE mindwell.dec_comments();

CREATE OR REPLACE FUNCTION mindwell.set_last_comment_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET last_comment = NEW.id 
        WHERE id = NEW.entry_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER last_comments_ins
    AFTER INSERT ON mindwell.comments
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_last_comment_ins();

CREATE OR REPLACE FUNCTION mindwell.set_last_comment_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.entries
        SET last_comment = (
            SELECT max(comments.id)
            FROM comments
            WHERE entry_id = OLD.entry_id AND id <> OLD.id
        )
        WHERE last_comment = OLD.id;
        
        RETURN OLD;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER last_comments_del
    BEFORE DELETE ON mindwell.comments
    FOR EACH ROW EXECUTE PROCEDURE mindwell.set_last_comment_del();



-- CREATE TABLE "comment_votes" --------------------------------
CREATE TABLE "mindwell"."comment_votes" (
	"user_id" Integer NOT NULL,
	"comment_id" Integer NOT NULL,
    "vote" Real NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "comment_vote_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "comment_vote_comment_id" FOREIGN KEY("comment_id") REFERENCES "mindwell"."comments"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_comment_vote" UNIQUE("user_id", "comment_id") );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_voted_comments" -------------------------
CREATE INDEX "index_voted_comments" ON "mindwell"."comment_votes" USING btree( "comment_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_comment_voted_users" --------------------
CREATE INDEX "index_comment_voted_users" ON "mindwell"."comment_votes" USING btree( "user_id" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.comment_votes_ins() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET up_votes = up_votes + (NEW.vote > 0)::int,
            down_votes = down_votes + (NEW.vote < 0)::int,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            rating = atan2(weight_sum + abs(NEW.vote), 2)
                * (vote_sum + NEW.vote) / (weight_sum + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.comment_id;
        
        WITH cmnt AS (
            SELECT author_id
            FROM mindwell.comments
            WHERE id = NEW.comment_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count + 1,
            vote_sum = vote_sum + NEW.vote,
            weight_sum = weight_sum + abs(NEW.vote),
            weight = atan2(vote_count + 1, 20) * (vote_sum + NEW.vote) 
                / (weight_sum + abs(NEW.vote)) / pi() * 2
        FROM cmnt
        WHERE user_id = cmnt.author_id 
            AND vote_weights.category = 
                (SELECT id FROM categories WHERE "type" = 'comment');

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.comment_votes_upd() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET up_votes = up_votes - (OLD.vote > 0)::int + (NEW.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int + (NEW.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            rating = atan2(weight_sum - abs(OLD.vote) + abs(NEW.vote), 2)
                * (vote_sum - OLD.vote + NEW.vote) / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 200
        WHERE id = NEW.comment_id;
        
        WITH cmnt AS (
            SELECT author_id
            FROM mindwell.comments
            WHERE id = NEW.comment_id
        )
        UPDATE mindwell.vote_weights
        SET vote_sum = vote_sum - OLD.vote + NEW.vote,
            weight_sum = weight_sum - abs(OLD.vote) + abs(NEW.vote),
            weight = atan2(vote_count, 20) * (vote_sum - OLD.vote + NEW.vote) 
                / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 2
        FROM cmnt
        WHERE user_id = cmnt.author_id
            AND vote_weights.category = 
                (SELECT id FROM categories WHERE "type" = 'comment');

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.comment_votes_del() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET up_votes = up_votes - (OLD.vote > 0)::int,
            down_votes = down_votes - (OLD.vote < 0)::int,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            rating = CASE WHEN weight_sum = abs(OLD.vote) THEN 0
                ELSE atan2(weight_sum - abs(OLD.vote), 2)
                    * (vote_sum - OLD.vote) / (weight_sum - abs(OLD.vote)) / pi() * 200
                END
        WHERE id = OLD.comment_id;
        
        WITH cmnt AS (
            SELECT author_id
            FROM mindwell.comments
            WHERE id = OLD.comment_id
        )
        UPDATE mindwell.vote_weights
        SET vote_count = vote_count - 1,
            vote_sum = vote_sum - OLD.vote,
            weight_sum = weight_sum - abs(OLD.vote),
            weight = CASE WHEN weight_sum = abs(OLD.vote) THEN 0.1
                ELSE atan2(vote_count - 1, 20) * (vote_sum - OLD.vote) 
                    / (weight_sum - abs(OLD.vote)) / pi() * 2
                END
        FROM cmnt
        WHERE user_id = cmnt.author_id
            AND vote_weights.category = 
                (SELECT id FROM categories WHERE "type" = 'comment');

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_comment_votes_ins
    AFTER INSERT ON mindwell.comment_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.comment_votes_ins();

CREATE TRIGGER cnt_comment_votes_upd
    AFTER UPDATE ON mindwell.comment_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.comment_votes_upd();

CREATE TRIGGER cnt_comment_votes_del
    AFTER DELETE ON mindwell.comment_votes
    FOR EACH ROW 
    EXECUTE PROCEDURE mindwell.comment_votes_del();

    

INSERT INTO mindwell.users
    (name, show_name, email, password_hash, api_key, invited_by)
    VALUES('Mindwell', 'Mindwell', '', '', '', 1);



-- CREATE TABLE "notification_type" ----------------------------
CREATE TABLE "mindwell"."notification_type" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."notification_type" VALUES(0, 'comment');
INSERT INTO "mindwell"."notification_type" VALUES(1, 'follower');
INSERT INTO "mindwell"."notification_type" VALUES(2, 'request');
INSERT INTO "mindwell"."notification_type" VALUES(3, 'accept');
INSERT INTO "mindwell"."notification_type" VALUES(4, 'invite');
INSERT INTO "mindwell"."notification_type" VALUES(5, 'welcome');
INSERT INTO "mindwell"."notification_type" VALUES(6, 'invited');
-- -------------------------------------------------------------



-- CREATE TABLE "notifications" --------------------------------
CREATE TABLE "mindwell"."notifications" (
    "id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "type" Integer NOT NULL,
    "subject_id" Integer NOT NULL,
    "read" Boolean DEFAULT FALSE NOT NULL,
	CONSTRAINT "unique_notification_id" PRIMARY KEY("id"),
    CONSTRAINT "notification_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "enum_notification_type" FOREIGN KEY("type") REFERENCES "mindwell"."notification_type"("id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_notification_id" ------------------------
CREATE UNIQUE INDEX "index_notification_id" ON "mindwell"."notifications" USING btree( "id" );
-- -------------------------------------------------------------

-- CREATE "index_notification_user_id" -------------------
CREATE INDEX "index_notification_user_id" ON "mindwell"."notifications" USING btree( "user_id" );
-- -------------------------------------------------------------



-- CREATE TABLE "images" ---------------------------------------
CREATE TABLE "mindwell"."images" (
	"id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
	"path" Text NOT NULL,
    "extension" Text NOT NULL,
    "processing" Boolean DEFAULT TRUE NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "image_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"));
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_image_id" -------------------------------
CREATE UNIQUE INDEX "index_image_id" ON "mindwell"."images" USING btree( "id" );
-- -------------------------------------------------------------



-- CREATE TABLE "size" -----------------------------------------
CREATE TABLE "mindwell"."size" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."size" VALUES(0, 'small');
INSERT INTO "mindwell"."size" VALUES(1, 'medium');
INSERT INTO "mindwell"."size" VALUES(2, 'large');
INSERT INTO "mindwell"."size" VALUES(3, 'thumbnail');
-- -------------------------------------------------------------



-- CREATE TABLE "image_sizes" ----------------------------------
CREATE TABLE "mindwell"."image_sizes" (
    "image_id" Integer NOT NULL,
    "size" Integer NOT NULL,
    "width" Integer NOT NULL,
    "height" Integer NOT NULL,
    CONSTRAINT "unique_image_size" PRIMARY KEY ("image_id", "size"),
    CONSTRAINT "unique_image_id" FOREIGN KEY ("image_id") REFERENCES "mindwell"."images"("id") ON DELETE CASCADE,
    CONSTRAINT "enum_image_size" FOREIGN KEY("size") REFERENCES "mindwell"."size"("id")
);
-- -------------------------------------------------------------

-- CREATE INDEX "index_image_size_id" --------------------------
CREATE INDEX "index_image_size_id" ON "mindwell"."image_sizes" USING btree( "image_id" );
-- -------------------------------------------------------------



-- CREATE TABLE "entry_images" ---------------------------------
CREATE TABLE "mindwell"."entry_images" (
    "entry_id" Integer NOT NULL,
    "image_id" Integer NOT NULL,
    CONSTRAINT "entry_images_entry" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE,
    CONSTRAINT "entry_images_image" FOREIGN KEY("image_id") REFERENCES "mindwell"."images"("id") ON DELETE CASCADE,
    CONSTRAINT "unique_entry_image" UNIQUE("entry_id", "image_id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_images_entry" ---------------------
CREATE INDEX "index_entry_images_entry" ON "mindwell"."entry_images" USING btree( "entry_id" );
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_images_image" ---------------------
CREATE INDEX "index_entry_images_image" ON "mindwell"."entry_images" USING btree( "image_id" );
-- -------------------------------------------------------------



-- CREATE TABLE "complain_type" -------------------------------
CREATE TABLE "mindwell"."complain_type" (
    "id" Integer UNIQUE NOT NULL,
    "type" Text NOT NULL );

INSERT INTO "mindwell"."complain_type" VALUES(0, 'comment');
INSERT INTO "mindwell"."complain_type" VALUES(1, 'entry');
-- -------------------------------------------------------------

-- CREATE TABLE "complains" ------------------------------------
CREATE TABLE "mindwell"."complains" (
    "id" Serial NOT NULL,
    "user_id" Integer NOT NULL,
    "created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "type" Integer NOT NULL,
    "subject_id" Integer NOT NULL,
    "content" Text NOT NULL,
    CONSTRAINT "unique_complain_id" PRIMARY KEY("id"),
    CONSTRAINT "complain_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "enum_complain_type" FOREIGN KEY("type") REFERENCES "mindwell"."complain_type"("id") );
;
-- -------------------------------------------------------------

-- CREATE "index_complain_user_id" -------------------
CREATE INDEX "index_complain_user_id" ON "mindwell"."complains" USING btree( "user_id" );
-- -------------------------------------------------------------



CREATE OR REPLACE FUNCTION mindwell.give_invites() RETURNS TABLE(user_id int) AS $$
    WITH inviters AS (
        UPDATE mindwell.users 
        SET last_invite = CURRENT_DATE
        WHERE ((id IN (
                    SELECT author_id 
                    FROM (
                        SELECT created_at, author_id, rating 
                        FROM entries 
                        WHERE age(created_at) <= interval '1 month' 
                            AND visible_for = (SELECT id FROM entry_privacy WHERE type = 'all')
                        ORDER BY rating DESC 
                        LIMIT 100) AS e
                    WHERE created_at::date = CURRENT_DATE - 3
                )
                AND (SELECT COUNT(*) FROM mindwell.invites WHERE referrer_id = users.id) < 3
            ) OR (
                last_invite = created_at::Date
                AND (
                    SELECT COUNT(DISTINCT entries.id)
                    FROM mindwell.entries, mindwell.entry_votes
                    WHERE entries.author_id = users.id AND entry_votes.entry_id = entries.id
                        AND entry_votes.vote > 0 AND entry_votes.user_id <> users.invited_by
                ) >= 10
            )) AND age(last_invite) >= interval '14 days'
            AND invite_ban <= CURRENT_DATE
            AND invited_by IS NOT NULL
        RETURNING users.id
    ), wc AS (
        SELECT COUNT(*) AS words FROM invite_words
    )
    INSERT INTO mindwell.invites(referrer_id, word1, word2, word3)
        SELECT inviters.id, 
            trunc(random() * wc.words),
            trunc(random() * wc.words),
            trunc(random() * wc.words)
        FROM inviters, wc
        ON CONFLICT (word1, word2, word3) DO NOTHING
    RETURNING referrer_id;
$$ LANGUAGE SQL;



CREATE OR REPLACE FUNCTION mindwell.recalc_karma() RETURNS VOID AS $$
    WITH upd AS (
        SELECT users.id, (users.karma * 4 + COALESCE(ek.karma, 0) + COALESCE(ck.karma, 0)) / 5 AS karma
        FROM mindwell.users
        LEFT JOIN (
            SELECT users.id, sum(entry_votes.vote) AS karma
            FROM mindwell.entry_votes, mindwell.entries, mindwell.users
            WHERE abs(entry_votes.vote) > 0.2 AND age(entries.created_at) <= interval '2 months'
                AND entry_votes.entry_id = entries.id AND entries.author_id = users.id
            GROUP BY users.id  
        ) as ek ON users.id = ek.id
        LEFT JOIN (
            SELECT users.id, sum(comment_votes.vote) / 10 AS karma
            FROM mindwell.comment_votes, mindwell.comments, mindwell.users
            WHERE abs(comment_votes.vote) > 0.2 AND age(comments.created_at) <= interval '2 months'
                AND comment_votes.comment_id = comments.id AND comments.author_id = users.id
            GROUP BY users.id  
        ) AS ck ON users.id = ck.id
    )
    UPDATE mindwell.users
    SET karma = upd.karma
    FROM upd
    WHERE users.id = upd.id;

    WITH upd AS (
        SELECT id, row_number() OVER (ORDER BY karma DESC, created_at ASC) as rank
        FROM mindwell.users
    )
    UPDATE mindwell.users
    SET rank = upd.rank
    FROM upd
    WHERE users.id = upd.id;
$$ LANGUAGE SQL;



CREATE OR REPLACE FUNCTION mindwell.can_view_tlog(user_id INTEGER, tlog TEXT) RETURNS BOOLEAN AS $$
    DECLARE
        tlog_id INTEGER;
    BEGIN
        SELECT users.id
        INTO tlog_id
		FROM users
		WHERE lower(users.name) = lower(tlog);

        RETURN (SELECT can_view_tlog(user_id, tlog_id));
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION mindwell.can_view_tlog(user_id INTEGER, tlog_id INTEGER) RETURNS BOOLEAN AS $$
    DECLARE
        privacy TEXT;
        is_ignored BOOLEAN;
        is_invited BOOLEAN;
        is_follower BOOLEAN;
    BEGIN
        IF user_id = tlog_id THEN
            RETURN TRUE;
        END IF;

        SELECT relation.type = 'ignored'
        INTO is_ignored
		FROM relations
		INNER JOIN relation ON relation.id = relations.type
		WHERE from_id = tlog_id AND to_id = user_id;

        IF is_ignored THEN
            RETURN FALSE;
        END IF;

        SELECT user_privacy.type
        INTO privacy
		FROM users
        LEFT JOIN user_privacy ON users.privacy = user_privacy.id
		WHERE users.id = tlog_id;

        CASE privacy
        WHEN 'all' THEN 
            RETURN TRUE;
        WHEN 'invited' THEN
            SELECT invited_by IS NOT NULL 
            INTO is_invited
            FROM users
            WHERE users.id = user_id;

            RETURN is_invited;
        WHEN 'followers' THEN
            SELECT relation.type = 'followed'
            INTO is_follower
            FROM relations
            INNER JOIN relation ON relation.id = relations.type
            WHERE from_id = user_id AND to_id = tlog_id;

            is_follower = COALESCE(is_follower, FALSE);
            RETURN is_follower;
        END CASE;

        RETURN FALSE;
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION mindwell.can_view_entry(user_id INTEGER, entry_id INTEGER) RETURNS BOOLEAN AS $$
    DECLARE
        author_id INTEGER;
        entry_privacy TEXT;
    BEGIN
		SELECT entries.author_id, entry_privacy.type
        INTO author_id, entry_privacy
		FROM entries
		INNER JOIN entry_privacy ON visible_for = entry_privacy.id
		WHERE entries.id = entry_id;

        RETURN (SELECT can_view_entry(user_id, entry_id, author_id, entry_privacy));
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION mindwell.can_view_entry(user_id INTEGER, entry_id INTEGER, author_id INTEGER, entry_privacy TEXT) RETURNS BOOLEAN AS $$
    DECLARE
        allowed BOOLEAN;
    BEGIN
        IF author_id = user_id THEN 
            RETURN TRUE;
        END IF;

        IF entry_privacy = 'anonymous' THEN
            RETURN TRUE;
        END IF;

        allowed = (SELECT can_view_tlog(user_id, author_id));

        IF NOT allowed THEN
            RETURN FALSE;
        END IF;

        CASE entry_privacy
        WHEN 'all' THEN
            RETURN TRUE;
        WHEN 'some' THEN
            SELECT TRUE
            INTO allowed
			FROM entries_privacy 
			WHERE entries_privacy.user_id = can_view_entry.user_id 
                AND entries_privacy.entry_id = can_view_entry.entry_id;
            
            allowed = COALESCE(allowed, FALSE);
            RETURN allowed;
        WHEN 'me' THEN
            RETURN FALSE;
        END CASE;

        RETURN FALSE;
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION mindwell.ban_user(userName TEXT) RETURNS TEXT AS $$
    BEGIN
        UPDATE mindwell.users
        SET api_key = (
            SELECT array_to_string(array(
                SELECT substr('abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789', 
                    trunc(random() * 62)::integer + 1, 1)
                FROM generate_series(1, 32)), '')
            ),
            password_hash = '', verified = false
        WHERE lower(users.name) = lower(userName);

        RETURN (SELECT name FROM mindwell.users WHERE id = (
            SELECT invited_by FROM users WHERE lower(name) = lower(userName)
        ));
    END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION mindwell.ban_invite(userName Text) RETURNS VOID AS $$
    DELETE FROM mindwell.invites
    WHERE referrer_id = (SELECT id FROM users WHERE lower(name) = lower(userName));

    UPDATE mindwell.users
    SET invite_ban = CURRENT_DATE + interval '1 month'
    WHERE lower(name) = lower(userName);
$$ LANGUAGE SQL;



CREATE OR REPLACE FUNCTION mindwell.delete_user(user_name TEXT) RETURNS VOID AS $$
    DECLARE
        del_id INTEGER;
    BEGIN
        del_id = (SELECT id FROM users WHERE lower(name) = lower(user_name));

        DELETE FROM mindwell.relations WHERE to_id = del_id;
        DELETE FROM mindwell.relations WHERE from_id = del_id;

        DELETE FROM mindwell.favorites WHERE favorites.user_id = del_id;
        DELETE FROM mindwell.watching WHERE watching.user_id = del_id;
        DELETE FROM mindwell.entries_privacy WHERE entries_privacy.user_id = del_id;
        
        DELETE FROM mindwell.entry_votes WHERE entry_votes.user_id = del_id;
        DELETE FROM mindwell.comment_votes WHERE comment_votes.user_id = del_id;
        DELETE FROM mindwell.vote_weights WHERE vote_weights.user_id = del_id;

        DELETE FROM mindwell.notifications
        WHERE notifications.user_id = del_id OR
            CASE (SELECT "type" FROM notification_type WHERE notification_type.id = notifications."type")
            WHEN 'comment' THEN
                (SELECT author_id FROM comments WHERE comments.id = notifications.subject_id) = del_id
            WHEN 'invite' THEN
                FALSE
            ELSE 
                notifications.subject_id = del_id
            END;

        DELETE FROM mindwell.images WHERE images.user_id = del_id;
        DELETE FROM mindwell.entries WHERE author_id = del_id;
        DELETE FROM mindwell.comments WHERE author_id = del_id;
        DELETE FROM mindwell.users WHERE id = del_id;
    END;
$$ LANGUAGE plpgsql;



-- -- CREATE TABLE "media" ----------------------------------------
-- CREATE TABLE "mindwell"."media" (
-- 	"id" Serial NOT NULL,
-- 	"duration" Integer NOT NULL,
-- 	"icon" Text NOT NULL,
-- 	"preview" Text NOT NULL,
-- 	"title" Text NOT NULL,
-- 	"url" Text NOT NULL,
-- 	"entry_id" Integer NOT NULL,
-- 	CONSTRAINT "unique_media_id" PRIMARY KEY( "id" ),
--     CONSTRAINT "media_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") );
--  ;
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_media_id" -------------------------------
-- CREATE INDEX "index_media_id" ON "mindwell"."media" USING btree( "id" );
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_media_entry" ----------------------------
-- CREATE INDEX "index_media_entry" ON "mindwell"."media" USING btree( "entry_id" );
-- -- -------------------------------------------------------------
--
--
--
-- -- CREATE TABLE "chats" ----------------------------------------
-- CREATE TABLE "mindwell"."chats" (
-- 	"id" Serial NOT NULL,
-- 	"messages_count" Integer DEFAULT 0 NOT NULL,
-- 	"avatar" Text DEFAULT '' NOT NULL,
-- 	CONSTRAINT "unique_chat_id" PRIMARY KEY( "id" ) );
--  ;
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_chat_id" --------------------------------
-- CREATE INDEX "index_chat_id" ON "mindwell"."chats" USING btree( "id" );
-- -- -------------------------------------------------------------
--
--
--
-- -- CREATE TABLE "messages" -------------------------------------
-- CREATE TABLE "mindwell"."messages" (
-- 	"id" Serial NOT NULL,
-- 	"chat_id" Integer NOT NULL,
-- 	"author_id" Integer NOT NULL,
-- 	"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
-- 	"content" Text NOT NULL,
-- 	"reply_to" Integer,
-- 	CONSTRAINT "unique_message_id" PRIMARY KEY( "id" ),
--     CONSTRAINT "message_user_id" FOREIGN KEY("author_id") REFERENCES "mindwell"."users"("id"),
--     CONSTRAINT "message_chat_id" FOREIGN KEY("chat_id") REFERENCES "mindwell"."chats"("id"),
--     CONSTRAINT "message_reply_to" FOREIGN KEY("reply_to") REFERENCES "mindwell"."comments"("id") );
--  ;
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_message_id" -----------------------------
-- CREATE INDEX "index_message_id" ON "mindwell"."messages" USING btree( "id" );
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_message_chat" ---------------------------
-- CREATE INDEX "index_message_chat" ON "mindwell"."messages" USING btree( "chat_id" );
-- -- -------------------------------------------------------------
--
-- CREATE OR REPLACE FUNCTION mindwell.count_messages() RETURNS TRIGGER AS $$
--     DECLARE
--         delta   integer;
--         chat_id integer;
--
--     BEGIN
--         IF (TG_OP = 'INSERT') THEN
--             delta = 1;
--             chat_id = NEW.chat_id;
--         ELSIF (TG_OP = 'DELETE') THEN
--             delta = -1;
--             chat_id = OLD.chat_id;
--         END IF;
--
--         UPDATE mindwell.chats
--         SET messages_count = messages_count + delta
--         WHERE id = chat_id;
--
--         RETURN NULL;
--     END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER cnt_messages
--     AFTER INSERT OR DELETE ON mindwell.messages
--     FOR EACH ROW EXECUTE PROCEDURE mindwell.count_messages();
--
--
--
-- -- CREATE TABLE "talker_status" --------------------------------
-- CREATE TABLE "mindwell"."talker_status" (
--     "id" Integer NOT NULL,
--     "type" Text NOT NULL );
--
-- INSERT INTO "mindwell"."talker_status" VALUES(0, "creator");
-- INSERT INTO "mindwell"."talker_status" VALUES(1, "banned");
-- INSERT INTO "mindwell"."talker_status" VALUES(2, "normal");
-- INSERT INTO "mindwell"."talker_status" VALUES(3, "left");
-- INSERT INTO "mindwell"."talker_status" VALUES(4, "admin");
-- -- -------------------------------------------------------------
--
--
--
-- -- CREATE TABLE "talking" --------------------------------------
-- CREATE TABLE "mindwell"."talking" (
-- 	"chat_id" Integer NOT NULL,
-- 	"last_read" Integer,
-- 	"user_id" Integer NOT NULL,
-- 	"unread_count" Text NOT NULL,
-- 	"status" Integer NOT NULL,
-- 	"not_disturb" Boolean DEFAULT false NOT NULL,
--     CONSTRAINT "talking_user_id" FOREIGN KEY("user_id") REFERENCES "mindwell"."users"("id"),
--     CONSTRAINT "talking_chat_id" FOREIGN KEY("chat_id") REFERENCES "mindwell"."chats"("id"),
--     CONSTRAINT "enum_talking_status" FOREIGN KEY("status") REFERENCES "mindwell"."talker_status"("id"));
--  ;
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_talking_chat" ---------------------------
-- CREATE INDEX "index_talking_chat" ON "mindwell"."talking" USING btree( "chat_id" );
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_talking_user" ---------------------------
-- CREATE INDEX "index_talking_user" ON "mindwell"."talking" USING btree( "user_id" )
--     WHERE "status" NOT IN (
--         SELECT "id" from "talker_status"
--         WHERE "type" = "banned" OR "type" = "left");
-- -- -------------------------------------------------------------
--
-- CREATE OR REPLACE FUNCTION mindwell.count_unread() RETURNS TRIGGER AS $$
--     BEGIN
--         IF (TG_OP = 'INSERT') THEN
--             UPDATE mindwell.talking
--             SET unread_count = unread_count + 1
--             WHERE talking.chat_id = NEW.chat_id AND talking.user_id <> NEW.user_id
--         ELSIF (TG_OP = 'DELETE') THEN
--             UPDATE mindwell.talking
--             SET unread_count = unread_count -1
--             WHERE talking.chat_id = OLD.chat_id AND talking.user_id <> OLD.user_id
--                 AND (last_read = NULL OR last_read < OLD.id)
--         END IF;
--
--         RETURN NULL;
--     END;
-- $$ LANGUAGE plpgsql;
--
-- CREATE TRIGGER cnt_unread
--     AFTER INSERT OR DELETE ON mindwell.messages
--     FOR EACH ROW EXECUTE PROCEDURE mindwell.count_unread();
