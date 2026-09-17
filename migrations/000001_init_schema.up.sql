CREATE TABLE events (
    events_id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    total_slots INT NOT NULL,
    available_slots INT NOT NULL CHECK (
        available_slots >= 0
        AND available_slots <= total_slots
    ),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE reservations (
    reservations_id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(events_id),
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    idempotency_key TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT uq_event_idempotency UNIQUE (event_id, idempotency_key)
);

CREATE INDEX idx_reservations_pending_expiry
    ON reservations (expires_at)
    WHERE status = 'pending';