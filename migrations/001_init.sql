CREATE TABLE IF NOT EXISTS menu_items (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
  emoji TEXT NOT NULL DEFAULT '🍽️',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS specials (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
  emoji TEXT NOT NULL DEFAULT '🔥',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stock_items (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  quantity NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  unit TEXT NOT NULL DEFAULT 'Kg',
  icon TEXT NOT NULL DEFAULT '📦',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
  id TEXT PRIMARY KEY,
  customer TEXT NOT NULL,
  date TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  items JSONB NOT NULL,
  total NUMERIC(12,2) NOT NULL CHECK (total >= 0),
  status TEXT NOT NULL DEFAULT 'Pending' CHECK (status IN ('Pending','Preparing','Delivered','Cancelled'))
);

INSERT INTO menu_items (name, description, price, emoji)
SELECT * FROM (VALUES
  ('Chicken Biryani', 'Aromatic basmati rice and tender chicken.', 280, '🍛'),
  ('Butter Chicken', 'Creamy tomato curry with rich spices.', 320, '🍗'),
  ('Mutton Rogan Josh', 'Slow-cooked mutton in Kashmiri spices.', 390, '🍖'),
  ('Tandoori Platter', 'Char-grilled favourites for sharing.', 450, '🍢')
) AS seed(name, description, price, emoji)
WHERE NOT EXISTS (SELECT 1 FROM menu_items);

INSERT INTO specials (name, description, price, emoji)
SELECT 'Chicken Biryani', 'Today only — aromatic basmati rice and tender chicken.', 280, '🍛'
WHERE NOT EXISTS (SELECT 1 FROM specials);

INSERT INTO stock_items (name, quantity, unit, icon)
SELECT * FROM (VALUES
  ('Chicken', 65, 'Kg', '🍗'),
  ('Basmati Rice', 42, 'Kg', '🍚'),
  ('Cooking Oil', 18, 'Litres', '🫗')
) AS seed(name, quantity, unit, icon)
WHERE NOT EXISTS (SELECT 1 FROM stock_items);
