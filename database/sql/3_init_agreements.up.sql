CREATE TABLE agreements (
  user_uuid uuid not null references users (uuid) on delete restrict on update restrict,
  document_name text not null,
  date timestamptz not null check (date > 'epoch'::timestamptz),

  primary key (user_uuid, document_name, date)
);

-- ensure a document exists with that name at that time
CREATE FUNCTION check_agreements_document() RETURNS TRIGGER AS $$
  BEGIN
    IF NOT EXISTS (SELECT 1 FROM documents WHERE name = NEW.document_name AND valid_from <= NEW.date) THEN
      RAISE EXCEPTION 'agreements_document_not_exist';
    END IF;
    RETURN NEW;
  END
$$ LANGUAGE plpgsql;
CREATE CONSTRAINT TRIGGER check_agreements_document_tgr
    AFTER INSERT ON agreements
    FOR EACH ROW
    EXECUTE PROCEDURE check_agreements_document();

-- make it impossible to update/delete agreements
CREATE FUNCTION check_agreements_immutable() RETURNS TRIGGER AS $$
  BEGIN
    RAISE EXCEPTION 'agreements_cannot_be_modified';
  END
$$ LANGUAGE plpgsql;
CREATE TRIGGER check_agreements_immutable_tgr
    BEFORE UPDATE OR DELETE ON agreements
    FOR EACH ROW
    EXECUTE PROCEDURE check_agreements_immutable();
