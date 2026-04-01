ALTER TABLE training_sessions
DROP CONSTRAINT IF EXISTS training_sessions_course_id_fkey;

ALTER TABLE training_sessions
ADD CONSTRAINT training_sessions_course_id_fkey
FOREIGN KEY (course_id) REFERENCES courses(id);

ALTER TABLE training_sessions
DROP CONSTRAINT IF EXISTS training_sessions_deck_id_fkey;

ALTER TABLE training_sessions
ADD CONSTRAINT training_sessions_deck_id_fkey
FOREIGN KEY (deck_id) REFERENCES decks(id);

ALTER TABLE training_answers
DROP CONSTRAINT IF EXISTS training_answers_card_id_fkey;

ALTER TABLE training_answers
ADD CONSTRAINT training_answers_card_id_fkey
FOREIGN KEY (card_id) REFERENCES cards(id);
