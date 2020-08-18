-- +goose Up
-- +goose StatementBegin
CREATE TABLE checks (
   id serial PRIMARY KEY,
   url VARCHAR NOT NULL,
   period VARCHAR NOT NULL,
   regexp VARCHAR NOT NULL,
   check_time TIMESTAMP,
   status VARCHAR NOT NULL,
   error_message VARCHAR NOT NULL,
   http_code INTEGER,
   duration INTEGER,
   created TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE checks;
-- +goose StatementEnd
