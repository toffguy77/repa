-- 010_question_tone.up.sql
-- A question's *tone*, and a group setting that restricts a group to the kind ones.
--
-- The bank shipped one set of 205 questions to a class of 14-year-olds and a flat share of 21-year-olds
-- alike. Tone lets the group choose, which matters for shareability (people share flattering things and
-- hide unflattering ones, and the shared card is the only organic acquisition channel) and for store
-- review, where anonymous-rating apps are judged on exactly this.

CREATE TYPE question_tone AS ENUM ('WARM', 'NEUTRAL', 'EDGY');

ALTER TABLE questions ADD COLUMN tone question_tone NOT NULL DEFAULT 'NEUTRAL';
ALTER TABLE groups ADD COLUMN kind_only BOOLEAN NOT NULL DEFAULT false;

-- Classify the rows already present, so an existing database is corrected without re-seeding. Mirrors
-- internal/service/questions/tone.go — edgy markers first, then warm, then the category fallback. The
-- two must be changed together; a test pins the expected distribution.
UPDATE questions SET tone = 'EDGY'
WHERE lower(text) ~ '(предаст|наябедничает|украд|съест чужую|сольёт|сольет|тайно|притворя|врёт|врет|обман|завид|бесит|сплетн|подставит|бросит|сдаст|спалит|подслушива|читает чужие|хуже всех|не умеет|позор|неудач|стыдн|осудит|высмеет|подколет|унизит|скандал|драк|истери|жад|лениво|лень|опазд)';

UPDATE questions SET tone = 'WARM'
WHERE tone <> 'EDGY'
  AND lower(text) ~ '(поможет|поддержит|выручит|обнимет|подбодрит|добр|щедр|надёжн|надежн|честн|вежлив|лучший друг|можно доверить|на кого можно положиться|заступится|утешит|защитит|первым поздравит|самый спокойн|самый весёл|самый весел|улыб|талант|умеет|лучше всех|научит|объяснит)';

-- Category fallback for anything no marker matched.
UPDATE questions SET tone = 'WARM'
WHERE tone = 'NEUTRAL' AND category = 'SKILLS'
  AND lower(text) !~ '(предаст|наябедничает|украд|хуже всех|не умеет)';

UPDATE questions SET tone = 'EDGY'
WHERE tone = 'NEUTRAL' AND category IN ('HOT', 'SECRETS');

CREATE INDEX idx_questions_tone ON questions(tone);
