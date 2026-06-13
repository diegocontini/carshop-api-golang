-- +goose Up
-- +goose StatementBegin

CREATE TABLE cars (
    id          BIGSERIAL PRIMARY KEY,
    new         BOOLEAN NOT NULL,
    brand       TEXT    NOT NULL,
    model       TEXT    NOT NULL,
    year        INTEGER NOT NULL,
    price       NUMERIC NOT NULL,
    color       TEXT    NOT NULL,
    km          INTEGER NOT NULL,
    description TEXT    NOT NULL
);

CREATE TABLE car_images (
    id     BIGSERIAL PRIMARY KEY,
    url    TEXT   NOT NULL,
    car_id BIGINT REFERENCES cars(id) ON DELETE CASCADE
);

CREATE INDEX ix_car_images_car_id ON car_images(car_id);

CREATE TABLE users (
    id                            BIGSERIAL PRIMARY KEY,
    username                      TEXT     NOT NULL UNIQUE,
    password                      TEXT     NOT NULL,
    email                         TEXT     NOT NULL,
    comission_per_sale_in_percent SMALLINT,
    role                          TEXT     NOT NULL CHECK (role IN ('admin', 'vendor'))
);

CREATE TABLE orders (
    id            BIGSERIAL PRIMARY KEY,
    customer_name TEXT        NOT NULL,
    order_date    TIMESTAMPTZ NOT NULL,
    total         NUMERIC     NOT NULL,
    vendor_id     BIGINT      NOT NULL REFERENCES users(id)
);

CREATE TABLE order_items (
    id       BIGSERIAL PRIMARY KEY,
    order_id BIGINT  NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    car_id   BIGINT  NOT NULL REFERENCES cars(id),
    price    NUMERIC NOT NULL,
    discount NUMERIC NOT NULL
);

CREATE INDEX ix_order_items_order_id ON order_items(order_id);

CREATE TABLE vendor_comissions (
    id                   BIGSERIAL PRIMARY KEY,
    vendor_id            BIGINT  NOT NULL REFERENCES users(id),
    vendor_name          TEXT    NOT NULL,
    comission_percentage NUMERIC NOT NULL,
    comission_amount     NUMERIC NOT NULL,
    order_id             BIGINT  NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    order_total          NUMERIC NOT NULL
);

CREATE INDEX ix_vendor_comissions_vendor_id ON vendor_comissions(vendor_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS vendor_comissions;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS car_images;
DROP TABLE IF EXISTS cars;

-- +goose StatementEnd
