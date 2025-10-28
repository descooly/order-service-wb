CREATE TABLE IF NOT EXISTS order_info (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR UNIQUE NOT NULL,
    track_number VARCHAR,
    entry VARCHAR,
    locale VARCHAR,
    internal_signature VARCHAR,
    customer_id VARCHAR,
    delivery_service VARCHAR,
    shardkey VARCHAR,
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard VARCHAR
);

CREATE TABLE IF NOT EXISTS delivery (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES order_info(id) ON DELETE CASCADE,
    d_name VARCHAR,
    phone VARCHAR,
    zip VARCHAR,
    city VARCHAR,
    address VARCHAR,
    region VARCHAR,
    email VARCHAR
);

CREATE TABLE IF NOT EXISTS payment (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES order_info(id) ON DELETE CASCADE,
    p_transaction VARCHAR,
    request_id VARCHAR,
    currency VARCHAR,
    provider VARCHAR,
    amount INTEGER,
    payment_dt INTEGER,
    bank VARCHAR,
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES order_info(id) ON DELETE CASCADE,
    chrt_id INTEGER,
    track_number VARCHAR,
    price INTEGER,
    rid VARCHAR,
    i_name VARCHAR,
    sale INTEGER,
    i_size VARCHAR,
    total_price INTEGER,
    nm_id INTEGER,
    brand VARCHAR,
    status INTEGER
);