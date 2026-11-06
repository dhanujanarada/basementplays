-- +goose Up
-- Core ticketing schema. The Go service owns everything in the `ticketing` schema.
-- Money is integer minor units (LKR cents); times are timestamptz in UTC.
-- Business rules that must survive bugs live here as CHECK constraints, with the
-- domain state machines in internal/domain as the first line of defence.

CREATE TABLE ticketing.seat_plans (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    key        text NOT NULL UNIQUE,
    name       text NOT NULL,
    layout     jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT seat_plans_key_not_blank CHECK (btrim(key) <> ''),
    CONSTRAINT seat_plans_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT seat_plans_layout_is_object CHECK (jsonb_typeof(layout) = 'object')
);

CREATE TABLE ticketing.shows (
    id              text PRIMARY KEY,
    production_slug text NOT NULL,
    title           text NOT NULL,
    starts_at       timestamptz NOT NULL,
    venue           text NOT NULL DEFAULT '',
    seat_plan_id    uuid REFERENCES ticketing.seat_plans (id) ON DELETE RESTRICT,
    status          text NOT NULL DEFAULT 'draft',
    currency        text NOT NULL DEFAULT 'LKR',
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT shows_id_not_blank CHECK (btrim(id) <> ''),
    CONSTRAINT shows_production_slug_not_blank CHECK (btrim(production_slug) <> ''),
    CONSTRAINT shows_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT shows_status_valid CHECK (status IN ('draft', 'on_sale', 'closed', 'cancelled')),
    CONSTRAINT shows_currency_valid CHECK (currency ~ '^[A-Z]{3}$')
);

CREATE INDEX shows_production_slug_idx ON ticketing.shows (production_slug);

CREATE TABLE ticketing.orders (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    show_id           text NOT NULL REFERENCES ticketing.shows (id) ON DELETE RESTRICT,
    status            text NOT NULL DEFAULT 'pending',
    source            text NOT NULL,
    payment_method    text,
    buyer_name        text,
    buyer_phone       text,
    buyer_email       text,
    total_cents       integer NOT NULL DEFAULT 0,
    access_token_hash text NOT NULL,
    idempotency_key   text,
    expires_at        timestamptz NOT NULL,
    paid_at           timestamptz,
    notes             text NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT orders_status_valid CHECK (
        status IN ('pending', 'paid', 'expired', 'cancelled', 'needs_review')
    ),
    CONSTRAINT orders_source_valid CHECK (source IN ('online', 'manual')),
    CONSTRAINT orders_payment_method_valid CHECK (
        payment_method IN ('payhere', 'bank_transfer', 'cash', 'whatsapp')
    ),
    CONSTRAINT orders_total_nonnegative CHECK (total_cents >= 0),
    CONSTRAINT orders_token_hash_not_blank CHECK (btrim(access_token_hash) <> ''),
    CONSTRAINT orders_paid_requires_paid_at CHECK (status <> 'paid' OR paid_at IS NOT NULL),
    -- idempotency keys are unique per show; NULL keys (manual orders) may repeat.
    UNIQUE (show_id, idempotency_key),
    -- Needed by the composite FKs on order_items and tickets.
    UNIQUE (id, show_id)
);

CREATE INDEX orders_status_expires_idx
    ON ticketing.orders (status, expires_at)
    WHERE status = 'pending';
CREATE INDEX orders_show_created_idx ON ticketing.orders (show_id, created_at);

CREATE TABLE ticketing.show_seats (
    show_id    text NOT NULL REFERENCES ticketing.shows (id) ON DELETE CASCADE,
    seat_code  text NOT NULL,
    tier_key   text NOT NULL,
    tier_name  text NOT NULL,
    label      text,
    price_cents integer NOT NULL,
    sort_index integer NOT NULL,
    status     text NOT NULL DEFAULT 'available',
    order_id   uuid,
    held_until timestamptz,
    PRIMARY KEY (show_id, seat_code),
    CONSTRAINT show_seats_code_not_blank CHECK (btrim(seat_code) <> ''),
    CONSTRAINT show_seats_tier_key_not_blank CHECK (btrim(tier_key) <> ''),
    CONSTRAINT show_seats_tier_name_not_blank CHECK (btrim(tier_name) <> ''),
    CONSTRAINT show_seats_price_nonnegative CHECK (price_cents >= 0),
    CONSTRAINT show_seats_sort_index_nonnegative CHECK (sort_index >= 0),
    CONSTRAINT show_seats_status_valid CHECK (
        status IN ('available', 'held', 'sold', 'blocked')
    ),
    CONSTRAINT show_seats_status_shape CHECK (
        (status = 'available' AND order_id IS NULL AND held_until IS NULL)
        OR (status = 'held' AND order_id IS NOT NULL AND held_until IS NOT NULL)
        OR (status = 'sold' AND order_id IS NOT NULL AND held_until IS NULL)
        OR (status = 'blocked' AND order_id IS NULL AND held_until IS NULL)
    )
);

CREATE INDEX show_seats_order_idx
    ON ticketing.show_seats (order_id)
    WHERE order_id IS NOT NULL;

CREATE TABLE ticketing.order_items (
    order_id    uuid NOT NULL,
    show_id     text NOT NULL,
    seat_code   text NOT NULL,
    price_cents integer NOT NULL,
    PRIMARY KEY (order_id, seat_code),
    CONSTRAINT order_items_price_nonnegative CHECK (price_cents >= 0),
    -- The order and the seat must belong to the same show.
    FOREIGN KEY (order_id, show_id)
        REFERENCES ticketing.orders (id, show_id) ON DELETE CASCADE,
    FOREIGN KEY (show_id, seat_code)
        REFERENCES ticketing.show_seats (show_id, seat_code) ON DELETE RESTRICT
);

CREATE TABLE ticketing.tickets (
    id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id  uuid NOT NULL,
    show_id   text NOT NULL,
    seat_code text NOT NULL,
    status    text NOT NULL DEFAULT 'issued',
    issued_at timestamptz NOT NULL DEFAULT now(),
    used_at   timestamptz,
    CONSTRAINT tickets_status_valid CHECK (status IN ('issued', 'used', 'void')),
    CONSTRAINT tickets_used_at_shape CHECK ((status = 'used') = (used_at IS NOT NULL)),
    FOREIGN KEY (order_id, show_id)
        REFERENCES ticketing.orders (id, show_id) ON DELETE CASCADE,
    FOREIGN KEY (show_id, seat_code)
        REFERENCES ticketing.show_seats (show_id, seat_code) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX one_live_ticket_per_seat
    ON ticketing.tickets (show_id, seat_code)
    WHERE status IN ('issued', 'used');
CREATE INDEX tickets_order_idx ON ticketing.tickets (order_id);

CREATE TABLE ticketing.payments (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            uuid NOT NULL REFERENCES ticketing.orders (id) ON DELETE RESTRICT,
    provider            text NOT NULL,
    provider_payment_id text,
    status_code         integer,
    amount_cents        integer NOT NULL,
    currency            text NOT NULL,
    raw                 jsonb NOT NULL DEFAULT '{}'::jsonb,
    received_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT payments_provider_not_blank CHECK (btrim(provider) <> ''),
    CONSTRAINT payments_amount_nonnegative CHECK (amount_cents >= 0),
    CONSTRAINT payments_currency_valid CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT payments_raw_is_object CHECK (jsonb_typeof(raw) = 'object'),
    -- NULL provider_payment_id values are distinct, so manual payments may repeat.
    UNIQUE (provider, provider_payment_id)
);

CREATE INDEX payments_order_idx ON ticketing.payments (order_id);

CREATE TABLE ticketing.audit_log (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    at        timestamptz NOT NULL DEFAULT now(),
    actor     text NOT NULL,
    action    text NOT NULL,
    entity    text NOT NULL,
    entity_id text NOT NULL,
    data      jsonb NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT audit_log_actor_not_blank CHECK (btrim(actor) <> ''),
    CONSTRAINT audit_log_action_not_blank CHECK (btrim(action) <> ''),
    CONSTRAINT audit_log_entity_not_blank CHECK (btrim(entity) <> ''),
    CONSTRAINT audit_log_data_is_object CHECK (jsonb_typeof(data) = 'object')
);

CREATE INDEX audit_log_entity_idx
    ON ticketing.audit_log (entity, entity_id, at DESC);

-- Added after orders exists: a held/sold seat must reference an order for the same
-- show. MATCH SIMPLE means NULL order_id rows (available/blocked) are not checked.
ALTER TABLE ticketing.show_seats
    ADD CONSTRAINT show_seats_order_fkey
    FOREIGN KEY (order_id, show_id)
        REFERENCES ticketing.orders (id, show_id) ON DELETE RESTRICT;

-- +goose Down
DROP TABLE IF EXISTS ticketing.audit_log;
DROP TABLE IF EXISTS ticketing.payments;
DROP TABLE IF EXISTS ticketing.tickets;
DROP TABLE IF EXISTS ticketing.order_items;
DROP TABLE IF EXISTS ticketing.show_seats;
DROP TABLE IF EXISTS ticketing.orders;
DROP TABLE IF EXISTS ticketing.shows;
DROP TABLE IF EXISTS ticketing.seat_plans;
