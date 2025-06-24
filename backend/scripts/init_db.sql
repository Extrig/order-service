-- Удаляем старые таблицы для повторной инициализации
DROP TABLE IF EXISTS items, payment, delivery, orders CASCADE;

-- Основная таблица заказов
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

-- Адрес доставки (1 к 1 с orders)
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

-- Информация об оплате (1 к 1 с orders)
CREATE TABLE payment (
                         order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid),
                         transaction TEXT,
                         request_id TEXT,
                         currency TEXT,
                         provider TEXT,
                         amount INT,
                         payment_dt BIGINT,
                         bank TEXT,
                         delivery_cost INT,
                         goods_total INT,
                         custom_fee INT
);

-- Товары в заказе (много к 1)
CREATE TABLE items (
                       id SERIAL PRIMARY KEY,
                       order_uid TEXT REFERENCES orders(order_uid),
                       chrt_id INT,
                       track_number TEXT,
                       price INT,
                       rid TEXT,
                       name TEXT,
                       sale INT,
                       size TEXT,
                       total_price INT,
                       nm_id INT,
                       brand TEXT,
                       status INT
);

-- Пользователь и права (если бы не создавался автоматически)
DO
$$
BEGIN
    IF NOT EXISTS (
        SELECT FROM pg_catalog.pg_roles WHERE rolname = 'order_user'
    ) THEN
        CREATE USER order_user WITH PASSWORD 'password';
END IF;
END
$$;

GRANT CONNECT ON DATABASE order_service TO order_user;
GRANT USAGE ON SCHEMA public TO order_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO order_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO order_user;
