CREATE TABLE persons (
  id           UUID,
  given_name   VARCHAR(128),
  family_name  VARCHAR(128),
  -- we allow extra space for the pre-formatted number
  phone        VARCHAR(14),
  email        VARCHAR(255) NOT NULL,
  backup_email VARCHAR(255),
  backup_phone VARCHAR(14),
  avatar_url   VARCHAR,

  CONSTRAINT persons_key PRIMARY KEY ( id ),
  CONSTRAINT persons_ref_users FOREIGN KEY ( id ) REFERENCES users ( id )
);

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

CREATE VIEW persons_join_users AS
  SELECT u.*,
      p.given_name, p.family_name, p.phone, p.email, p.backup_email, p.backup_phone, p.avatar_url
    FROM persons p JOIN users_join_entity u ON p.id=u.id;
