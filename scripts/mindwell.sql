CREATE SCHEMA "mindwell";

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
	"karma" Real DEFAULT 0 NOT NULL,
	"created_at" Timestamp With Time Zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "invited_by" Integer NOT NULL,
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

-- CREATE INDEX "index_invited_by" -----------------------------
CREATE INDEX "index_invited_by" ON "mindwell"."users" USING btree( "invited_by" );
-- -------------------------------------------------------------

CREATE OR REPLACE FUNCTION mindwell.count_invited() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.users
        SET invited_count = invited_count + 1 
        WHERE id = NEW.invited_by;

        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_invited
    AFTER INSERT ON mindwell.users
    FOR EACH ROW EXECUTE PROCEDURE mindwell.count_invited();



CREATE VIEW mindwell.short_users AS
SELECT id, name, show_name,
    now() - last_seen_at < interval '15 minutes' AS is_online,
    avatar
FROM mindwell.users;



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
    users.karma,
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
    users.country,
    users.city,
    users.email,
    users.verified,
    users.api_key,
    users.valid_thru,
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
FROM mindwell.users, mindwell.short_users,
    mindwell.gender, mindwell.user_privacy, mindwell.font_family, mindwell.alignment
WHERE users.invited_by = short_users.id
    AND users.gender = gender.id
    AND users.privacy = user_privacy.id
    AND users.font_family = font_family.id
    AND users.text_alignment = alignment.id;

    

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



-- CREATE TABLE "invites" ------------------------------------
CREATE TABLE "mindwell"."invites" (
    "id" Serial NOT NULL,
    "referrer_id" Integer NOT NULL,
    "word1" Integer NOT NULL,
    "word2" Integer NOT NULL,
    "word3" Integer NOT NULL,
    CONSTRAINT "unique_invite_id" PRIMARY KEY( "id" ),
    CONSTRAINT "invite_word1" FOREIGN KEY("word1") REFERENCES "mindwell"."invite_words"("id"),
    CONSTRAINT "invite_word2" FOREIGN KEY("word2") REFERENCES "mindwell"."invite_words"("id"),
    CONSTRAINT "invite_word3" FOREIGN KEY("word3") REFERENCES "mindwell"."invite_words"("id") );
;
-- -------------------------------------------------------------

-- CREATE INDEX "index_referrer_id" -------------------------
CREATE INDEX "index_referrer_id" ON "mindwell"."invites" USING btree( "referrer_id" );
-- -------------------------------------------------------------

INSERT INTO mindwell.invites (referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1);
INSERT INTO mindwell.invites (referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1);
INSERT INTO mindwell.invites (referrer_id, word1, word2, word3) VALUES(1, 1, 1, 1);

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
	"rating" Real DEFAULT 0 NOT NULL,
    "up_votes" Integer DEFAULT 0 NOT NULL,
    "down_votes" Integer DEFAULT 0 NOT NULL,
    "vote_sum" Real DEFAULT 0 NOT NULL,
    "weight_sum" Real DEFAULT 0 NOT NULL,
    "category" Integer NOT NULL,
	"comments_count" Integer DEFAULT 0 NOT NULL,
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
            rating = atan(weight_sum + abs(NEW.vote))
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
            weight = atan2(vote_count + 1, 5) * (vote_sum + NEW.vote) 
                / (weight_sum + abs(NEW.vote)) / pi() * 2
        FROM entry
        WHERE user_id = entry.author_id 
            AND vote_weights.category = entry.category;

        IF abs(NEW.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = NEW.entry_id
            )
            UPDATE mindwell.users
            SET karma = karma + NEW.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

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
            rating = atan(weight_sum - abs(OLD.vote) + abs(NEW.vote))
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
            weight = atan2(vote_count, 5) * (vote_sum - OLD.vote + NEW.vote) 
                / (weight_sum - abs(OLD.vote) + abs(NEW.vote)) / pi() * 2
        FROM entry
        WHERE user_id = entry.author_id
            AND vote_weights.category = entry.category;

        IF abs(OLD.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = OLD.entry_id
            )
            UPDATE mindwell.users
            SET karma = karma - OLD.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

        IF abs(NEW.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = NEW.entry_id
            )
            UPDATE mindwell.users
            SET karma = karma + NEW.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

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
                ELSE atan(weight_sum - abs(OLD.vote))
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
                ELSE atan2(vote_count - 1, 5) * (vote_sum - OLD.vote) 
                    / (weight_sum - abs(OLD.vote)) / pi() * 2
                END
        FROM entry
        WHERE user_id = entry.author_id
            AND vote_weights.category = entry.category;

        IF abs(OLD.vote) > 0.2 THEN
            WITH entry AS (
                SELECT author_id
                FROM mindwell.entries
                WHERE id = OLD.entry_id
            )
            UPDATE mindwell.users
            SET karma = karma - OLD.vote * 5
            FROM entry
            WHERE users.id = entry.author_id;
        END IF;

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
	"rating" Integer DEFAULT 0 NOT NULL,
	CONSTRAINT "unique_comment_id" PRIMARY KEY( "id" ),
    CONSTRAINT "comment_user_id" FOREIGN KEY("author_id") REFERENCES "mindwell"."users"("id"),
    CONSTRAINT "comment_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") ON DELETE CASCADE );
 ;
-- -------------------------------------------------------------

-- CREATE INDEX "index_entry_id" -------------------------------
CREATE INDEX "index_comment_entry_id" ON "mindwell"."comments" USING btree( "entry_id" );
-- -------------------------------------------------------------

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



-- CREATE TABLE "comment_votes" --------------------------------
CREATE TABLE "mindwell"."comment_votes" (
	"user_id" Integer NOT NULL,
	"comment_id" Integer NOT NULL,
	"positive" Boolean NOT NULL,
	"taken" Boolean DEFAULT TRUE NOT NULL,
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

CREATE OR REPLACE FUNCTION mindwell.inc_comment_votes() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET rating = rating + 1
        WHERE id = NEW.comment_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.dec_comment_votes() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET rating = rating - 1
        WHERE id = OLD.comment_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.inc_comment_votes2() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET rating = rating + 2
        WHERE id = NEW.comment_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION mindwell.dec_comment_votes2() RETURNS TRIGGER AS $$
    BEGIN
        UPDATE mindwell.comments
        SET rating = rating - 2
        WHERE id = OLD.comment_id;
        
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cnt_comment_votes_ins_inc
    AFTER INSERT ON mindwell.comment_votes
    FOR EACH ROW 
    WHEN (NEW."positive" = true)
    EXECUTE PROCEDURE mindwell.inc_comment_votes();

CREATE TRIGGER cnt_comment_votes_ins_dec
    AFTER INSERT ON mindwell.comment_votes
    FOR EACH ROW 
    WHEN (NEW."positive" = false)
    EXECUTE PROCEDURE mindwell.dec_comment_votes();

CREATE TRIGGER cnt_comment_votes_upd_inc
    AFTER UPDATE ON mindwell.comment_votes
    FOR EACH ROW 
    WHEN (NEW."positive" = true)
    EXECUTE PROCEDURE mindwell.inc_comment_votes2();

CREATE TRIGGER cnt_comment_votes_upd_dec
    AFTER UPDATE ON mindwell.comment_votes
    FOR EACH ROW 
    WHEN (NEW."positive" = false)
    EXECUTE PROCEDURE mindwell.dec_comment_votes2();

CREATE TRIGGER cnt_comment_votes_del_dec
    AFTER DELETE ON mindwell.comment_votes
    FOR EACH ROW 
    WHEN (OLD."positive" = true)
    EXECUTE PROCEDURE mindwell.dec_comment_votes();

CREATE TRIGGER cnt_comment_votes_del_inc
    AFTER DELETE ON mindwell.comment_votes
    FOR EACH ROW 
    WHEN (OLD."positive" = false)
    EXECUTE PROCEDURE mindwell.inc_comment_votes();

    

INSERT INTO mindwell.users
    (name, show_name, email, password_hash, api_key, invited_by)
    VALUES('Mindwell', 'Mindwell', '', '', '', 1);

-- -- CREATE TABLE "images" ---------------------------------------
-- CREATE TABLE "mindwell"."images" (
-- 	"id" Serial NOT NULL,
-- 	"entry_id" Integer NOT NULL,
-- 	"url" Text NOT NULL,
-- 	"sorting" SmallInt NOT NULL,
-- 	CONSTRAINT "unique_image_id" PRIMARY KEY( "id" ),
--     CONSTRAINT "image_entry_id" FOREIGN KEY("entry_id") REFERENCES "mindwell"."entries"("id") );
--  ;
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_image_id" -------------------------------
-- CREATE INDEX "index_image_id" ON "mindwell"."images" USING btree( "id" );
-- -- -------------------------------------------------------------
--
-- -- CREATE INDEX "index_image_entry" ----------------------------
-- CREATE INDEX "index_image_entry" ON "mindwell"."images" USING btree( "entry_id" );
-- -- -------------------------------------------------------------
--
--
--
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
