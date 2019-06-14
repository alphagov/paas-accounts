ALTER TABLE users DROP CONSTRAINT users_email_key;
ALTER TABLE users ADD COLUMN username text UNIQUE;
