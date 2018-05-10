CREATE TABLE documents (
  name text not null check (length(name) > 0),
  content text not null check (length(content) > 0),
  valid_from timestamptz not null check (valid_from > 'epoch'::timestamptz),

  primary key (name, valid_from)
);

-- trigger to ensure you cannot insert content into a document's history
CREATE FUNCTION check_linear_document_history() RETURNS TRIGGER AS $$
  BEGIN
    IF EXISTS (SELECT 1 FROM documents WHERE name = NEW.name AND valid_from >= NEW.valid_from) THEN
      RAISE EXCEPTION 'cannot_alter_document_history';
    END IF;
    RETURN NEW;
  END
$$ LANGUAGE plpgsql;
CREATE TRIGGER check_linear_document_history_tgr
    BEFORE INSERT ON documents
    FOR EACH ROW
    EXECUTE PROCEDURE check_linear_document_history();

-- make it impossible to update/delete documents
CREATE FUNCTION check_documents_immutable() RETURNS TRIGGER AS $$
  BEGIN
    RAISE EXCEPTION 'documents_cannot_be_modified';
  END
$$ LANGUAGE plpgsql;
CREATE TRIGGER check_documents_immutable_tgr
    BEFORE UPDATE OR DELETE ON documents
    FOR EACH ROW
    EXECUTE PROCEDURE check_documents_immutable();
