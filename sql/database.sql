CREATE DATABASE twee;

\c twee;

CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    username VARCHAR NOT NULL,
    hash_password VARCHAR NOT NULL,
    followee_count INTEGER DEFAULT 0,
    follower_count INTEGER DEFAULT 0
);

CREATE TABLE tweets(
    id SERIAL PRIMARY KEY,
    content VARCHAR NOT NULL,
    created_at TIMESTAMP with time zone DEFAULT now(),
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE follows(
    id SERIAL PRIMARY KEY,
    followee INTEGER NOT NULL,
    follower INTEGER NOT NULL,
    FOREIGN KEY (followee) REFERENCES users (id),
    FOREIGN KEY (follower) REFERENCES users (id)
);