# paas-accounts

API for storing information about PaaS users.

## Build

```
go build -o paas-accounts .
```

## Run

```
BASIC_AUTH_USERNAME="some-user" \
BASIC_AUTH_PASSWORD="some-pass" \
DATABASE_URL="postgres://..." \
PORT=8080 \
	./paas-accounts
```

## Deploy

A manifest.yml exists for deploying to cloudfoundry. You should ensure the required environment variables are in place and that a suitable postgres database service is bound.

## API

### PUT /documents/:name

Create or update a document:

    curl -u <USER>:<PASS> -H "Content-Type: application/json" -X PUT -d '{"content": "my content"}' https://<HOSTNAME>/documents/my_document

### GET /documents/:name

Retrieve an existing document:

    curl -u <USER>:<PASS> https://<HOSTNAME>/documents/my_document

## Agreements

### POST /agreements

Record a user's agreement to a document:

    curl -u <USER>:<PASS> -H "Content-Type: application/json" -X POST -d '{"user_uuid": "00000000-0000-0000-0000-000000000001", "document_name": "my_document"}' https://<HOSTNAME>/agreements

### GET /users/:uuid/documents

Get all documents for a user:

    curl -u <USER>:<PASS> https://<HOSTNAME>/users/00000000-0000-0000-0000-000000000001/documents

Get all documents for a user that need agreement:

    curl -u <USER>:<PASS> -G -d agreed=false https://<HOSTNAME>/users/00000000-0000-0000-0000-000000000001/documents
