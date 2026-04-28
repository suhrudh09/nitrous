-- Nitrous PostgreSQL schema
-- This schema matches the current backend models and in-memory collections.
-- It is designed to be the first step before rewriting handlers to use SQL.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    icon TEXT NOT NULL,
    live_count INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    date TIMESTAMPTZ NOT NULL,
    time TEXT,
    is_live BOOLEAN NOT NULL DEFAULT FALSE,
    category TEXT NOT NULL,
    thumbnail_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_events_category
        FOREIGN KEY (category) REFERENCES categories(slug)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);
CREATE INDEX IF NOT EXISTS idx_events_is_live ON events(is_live);
CREATE INDEX IF NOT EXISTS idx_events_date ON events(date);

CREATE TABLE IF NOT EXISTS journeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL,
    badge TEXT,
    slots_left INTEGER NOT NULL DEFAULT 0,
    date TIMESTAMPTZ NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    thumbnail_url TEXT
);

CREATE INDEX IF NOT EXISTS idx_journeys_category ON journeys(category);
CREATE INDEX IF NOT EXISTS idx_journeys_date ON journeys(date);

CREATE TABLE IF NOT EXISTS merch_items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    category TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_merch_items_category ON merch_items(category);

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    country TEXT,
    followers_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_drivers (
    team_id UUID NOT NULL,
    driver_name TEXT NOT NULL,
    PRIMARY KEY (team_id, driver_name),
    CONSTRAINT fk_team_drivers_team
        FOREIGN KEY (team_id) REFERENCES teams(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS team_followers (
    team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id),
    CONSTRAINT fk_team_followers_team
        FOREIGN KEY (team_id) REFERENCES teams(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT fk_team_followers_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reminders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    event_id UUID NOT NULL,
    message TEXT,
    remind_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_reminders_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT fk_reminders_event
        FOREIGN KEY (event_id) REFERENCES events(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reminders_user_id ON reminders(user_id);
CREATE INDEX IF NOT EXISTS idx_reminders_event_id ON reminders(event_id);
CREATE INDEX IF NOT EXISTS idx_reminders_remind_at ON reminders(remind_at);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total NUMERIC(12, 2) NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL,
    merch_item_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12, 2) NOT NULL CHECK (unit_price > 0),
    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id) REFERENCES orders(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT fk_order_items_merch_item
        FOREIGN KEY (merch_item_id) REFERENCES merch_items(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_merch_item_id ON order_items(merch_item_id);

-- Garage user configurations
CREATE TABLE IF NOT EXISTS garage_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    make TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER NOT NULL,
    engine TEXT NOT NULL,
    tuning TEXT NOT NULL DEFAULT 'stock',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_garage_configs_user_id ON garage_configs(user_id);

-- Payment transactions
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(12, 2) NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    status TEXT NOT NULL DEFAULT 'pending',
    payment_method TEXT NOT NULL,
    stripe_payment_intent_id TEXT,
    stripe_customer_id TEXT,
    description TEXT,
    reference_type TEXT,
    reference_id TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_stripe_id ON payments(stripe_payment_intent_id);
CREATE INDEX IF NOT EXISTS idx_payments_reference ON payments(reference_type, reference_id);

CREATE TABLE IF NOT EXISTS passes (
    id TEXT PRIMARY KEY,
    tier TEXT NOT NULL,
    event_name TEXT NOT NULL,
    location TEXT NOT NULL,
    event_date TIMESTAMPTZ NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    perks JSONB NOT NULL DEFAULT '[]'::jsonb,
    spots_left INTEGER NOT NULL DEFAULT 0,
    total_spots INTEGER NOT NULL DEFAULT 0,
    badge TEXT,
    tier_color TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_passes_category ON passes(category);
CREATE INDEX IF NOT EXISTS idx_passes_event_date ON passes(event_date);

CREATE TABLE IF NOT EXISTS pass_purchases (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    pass_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_pass_purchases_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT fk_pass_purchases_pass
        FOREIGN KEY (pass_id) REFERENCES passes(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT uq_pass_purchase_once UNIQUE (user_id, pass_id)
);

CREATE INDEX IF NOT EXISTS idx_pass_purchases_user_id ON pass_purchases(user_id);
CREATE INDEX IF NOT EXISTS idx_pass_purchases_pass_id ON pass_purchases(pass_id);
