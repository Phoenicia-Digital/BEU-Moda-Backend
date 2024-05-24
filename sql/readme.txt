There are two Tables that should exit for the USER

Table users:

CREATE TABLE users (
  uid SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL  -- Hashed password (security best practice)
);

Table billing_info:

CREATE TABLE billing_info (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(uid),  -- Foreign key to users.uid
  country VARCHAR(255),
  province VARCHAR(255),
  city VARCHAR(255),
  street VARCHAR(255),
  building VARCHAR(255),
  floor VARCHAR(255),
  phone_number INTEGER,
  first_name VARCHAR(255),
  last_name VARCHAR(255)
);