CREATE DATABASE tweets;

\c tweets;

CREATE TABLE tweets(
    id SERIAL PRIMARY KEY,
    content VARCHAR NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP with time zone DEFAULT now()
);