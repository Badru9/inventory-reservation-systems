CREATE TABLE IF NOT EXISTS items (
  item_id      TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  total_stock  INTEGER NOT NULL CHECK (total_stock >= 0),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reservations (
  reservation_id  TEXT PRIMARY KEY,
  user_id         TEXT NOT NULL,
  item_id         TEXT NOT NULL REFERENCES items(item_id),
  quantity        INTEGER NOT NULL CHECK (quantity > 0),
  status          TEXT NOT NULL CHECK (status IN ('active','confirmed','expired')),
  expires_at      TIMESTAMPTZ NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  confirmed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_reservations_item_status ON reservations(item_id, status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires     ON reservations(expires_at) WHERE status = 'active';

INSERT INTO items(item_id, name, total_stock) VALUES
  ('item_4021', 'Flash Sale Headphones', 100),
  ('item_7777', 'Limited Sneakers',      50)
ON CONFLICT (item_id) DO NOTHING;
