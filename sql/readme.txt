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

Table sessions:

CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,
    user_uid INTEGER NOT NULL REFERENCES users(uid),
    login_time TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

Table admins:

CREATE TABLE admins (
  uid SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL  -- Hashed password (security best practice)
);

Table admin_sessions:

CREATE TABLE admin_sessions (
    id SERIAL PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,
    user_uid INTEGER NOT NULL REFERENCES admins(uid),
    login_time TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

Table items:

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    quantity SMALLINT NOT NULL,
    color VARCHAR(50)
);

Table pending_orders:

CREATE TABLE pending_orders (
    order_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(uid),
    ordered_items jsonb NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    order_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

