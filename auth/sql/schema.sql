CREATE DATABASE auth;

\c auth;

CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    username VARCHAR NOT NULL,
    hash_password VARCHAR NOT NULL
);