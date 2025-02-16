CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    pass_hash BYTEA NOT NULL,
    coin_balance INTEGER NOT NULL DEFAULT 1000, -- каждому новому сотруднику выдаётся 1000 монет
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

CREATE TABLE IF NOT EXISTS merch (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    price INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE  -- поле для soft deletion
);

CREATE INDEX IF NOT EXISTS idx_merch_name ON merch (name);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merch_id INTEGER NOT NULL REFERENCES merch(id),
    quantity INTEGER NOT NULL,
    total_price INTEGER NOT NULL,  -- вычисляется как quantity * price
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);

CREATE TABLE IF NOT EXISTS coin_transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    type TEXT NOT NULL,       -- тип операции: 'purchase', 'transfer', 'gift', 'credit' и т.п.
    related_user_id INTEGER,  -- если перевод между сотрудниками, id другого пользователя (может быть NULL)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coin_tx_user_id ON coin_transactions (user_id);