BEGIN TRANSACTION;

ALTER TABLE urls DROP CONSTRAINT urls_original_url_key;

ALTER TABLE urls ADD COLUMN user_uuid UUID NOT NULL;

ALTER TABLE urls ADD CONSTRAINT urls_user_id_original_url_uq UNIQUE (user_uuid, original_url);

COMMIT;