BEGIN TRANSACTION;

ALTER TABLE urls DROP CONSTRAINT urls_user_id_original_url_uq;

ALTER TABLE urls DROP COLUMN user_uuid;

ALTER TABLE urls ADD CONSTRAINT urls_original_url_key UNIQUE (original_url);

COMMIT;