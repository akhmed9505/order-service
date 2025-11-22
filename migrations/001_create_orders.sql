-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard TEXT
);

CREATE INDEX idx_orders_track_number ON orders(track_number);
CREATE INDEX idx_orders_customer_id ON orders(customer_id);

CREATE TABLE delivery (
    order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid),
    name TEXT,
    phone TEXT,
    zip TEXT,
    city TEXT,
    address TEXT,
    region TEXT,
    email TEXT
);

CREATE TABLE payment (
    order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid),
    transaction TEXT,
    request_id TEXT,
    currency TEXT,
    provider TEXT,
    amount BIGINT,
    payment_dt BIGINT,
    bank TEXT,
    delivery_cost BIGINT,
    goods_total BIGINT,
    custom_fee BIGINT
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid TEXT REFERENCES orders(order_uid),
    chrt_id BIGINT,
    track_number TEXT,
    price BIGINT,
    rid TEXT,
    name TEXT,
    sale INT,
    size TEXT,
    total_price BIGINT,
    nm_id BIGINT,
    brand TEXT,
    status INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS delivery;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
