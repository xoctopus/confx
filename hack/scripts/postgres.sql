CREATE DATABASE test;

\c test;

CREATE TABLE x (datetime timestamp(3), timestamp bigint);

INSERT INTO x (datetime,timestamp) VALUES ('1988-10-24 07:00:00.123', 593650800123);
