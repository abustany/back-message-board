# Back Message Board

[![Build Status](https://travis-ci.com/abustany/back-message-board.svg?branch=master)](https://travis-ci.com/abustany/back-message-board)

This is a small web server exposing a REST API allowing to post and review
messages. Posting messages is allowed for any user without authentication, while
listing, reading and modifying messages is allowed for the administrator only.

## Compiling

Run `make` to compile the server, called `server`.

## Running tests

Run `make test` to run the tests.

## REST API

### Data types

#### Post

```
{
  // Unique ID of this message, assigned by the server when the message is
  // created.
  id: String,

  // Name and email of the message author
  author: String,
  email: String,

  // Creation time of the message, assigned by the server when the message is
  // created. In RFC3339 format.
  created: String,

  // Content of the message
  message: String
}
```

#### ListResponse

```
{
  // List of posts on this page
  posts: Post[],

  // Cursor to be used at the next list call to continue listing. If not set,
  // this is the last page.
  next: String
}
```

### Endpoints

#### POST /post

Authentication required: no
Request body: A JSON encoded `Message` object
Reply: an HTTP 201 if the post was created, an HTTP error status else

Saves a new post in the store.

#### GET /admin/posts?n=N&cursor=CURSOR

Authentication required: yes
Query parameters:

- `n`: Desired number of results per page
- `cursor`: Used for pagination. Not set for the first page, set to the value of
  the `next` from the previous `ListResponse` for subsequent ones.

Reply: a `ListResponse` object with the results

Lists the posts in the store.

#### GET /admin/posts/ID

Authentication required: yes
URL parameters:

- ID: ID of the post to retrieve

Reply: a `Message` object, or a HTTP 404 if no such ID exists in the store

Retrieves a single post from the store.

#### POST /admin/posts

Authenticaton required: yes
Request body: a JSON encoded `Message` object
Reply: an HTTP 200 if the update succeeded, an HTTP error status else

Updates a post in the store. The post ID is read from the post in the request
body. Updating a non existing post is an error. All fields of a post can be
updated (including its creation time), except its ID.

Partial updates are supported by setting only the relevant fields in the
`Message` object, for example, to update the `author` field of the post with ID
`ID` while leaving the other fields untouched, use the following:

```json
{"id": "ID", "author": "new value"}
```

## Loading data at startup

The `-loadCSV` command line flag allows populating the messages from a CSV file
on startup. The CSV records in the file should have five fields:

- ID: unique ID of that message
- Name: name of the message author
- Email: email of the message author
- Message: contents of the message
- Created: creation date of the message, in RFC3339 format

The first record in the CSV file is considered as a header, and is skipped.

## Docker image

The repository provides a Dockerfile for the server, the resulting Docker image
uses the following environment variables to configure itself:

- `LISTEN_ADDRESS`: Sets the address the server listens on (by default `0.0.0.0:1412`)
- `ADMIN_USER` and `ADMIN_PASSWORD`: Sets the username and passwor for the admin
  user. If not set, a password will be auto generated on each start, and printed
  on the console.
- `LOAD_CSV`: If provided, path to a CSV file that should be loaded on startup.

For example, if you have a file `/tmp/messages.csv` with some data, and want to
have an admin user called `admin` with a password `s3cr3t`, you would run:

```
docker build -t back-message-board:latest .

docker run \
  -e ADMIN_USER=admin \
  -e ADMIN_PASSWORD=s3cr3t \
  -e LOAD_CSV=/tmp/messages.csv \
  -v /tmp/messages.csv:/tmp/messages.csv \
  back-message-board:latest
```

## Limitations

The data is at the moment stored in memory only, although adding new storage
option is only a matter of extending the `poststore.Store` interface.
