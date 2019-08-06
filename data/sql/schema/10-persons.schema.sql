CREATE TABLE persons (
  id BIGINT,
  given_name VARCHAR(128),
  family_name VARCHAR(128),
  -- we allow extra space for the pre-formatted number
  phone VARCHAR(14),
  email VARCHAR(255) NOT NULL,
  backup_email VARCHAR(255),
  backup_phone VARCHAR(14),
  CONSTRAINT persons_key PRIMARY KEY ( id ),
  CONSTRAINT persons_ref_users FOREIGN KEY ( id ) REFERENCES users ( id )
);

-- Postgres
CREATE OR REPLACE FUNCTION trigger_persons_phone_format()
  RETURNS TRIGGER AS '
BEGIN
  NEW.phone=NUMERIC_ONLY(NEW.phone);
  NEW.backup_phone=NUMERIC_ONLY(NEW.backup_phone);
  RETURN NEW;
END' LANGUAGE 'plpgsql';

CREATE TRIGGER persons_phone_format
  BEFORE INSERT OR UPDATE ON persons
  FOR EACH ROW
  EXECUTE PROCEDURE trigger_persons_phone_format();

-- MySQL
-- DELIMITER //
-- CREATE TRIGGER persons_phone_format
--  BEFORE INSERT ON persons FOR EACH ROW
--    BEGIN
--       SET new.phone=(SELECT NUMERIC_ONLY(new.phone));
--       SET new.phone_backup=(SELECT NUMERIC_ONLY(new.phone_backup));
--     END;//
-- DELIMITER ;
