CREATE TABLE papers (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  name VARCHAR(255) NOT NULL,
  cost_per_square_inch REAL NOT NULL,
  finish TEXT CHECK ( finish in ('glossy', 'matte', 'luster') ) NOT NULL
);

CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  is_admin BOOLEAN NOT NULL,
  created DATETIME NOT NULL
);

CREATE TABLE pictures (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  name VARCHAR(255) NOT NULL,
  user_id INTEGER NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE shipping_profiles (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  name TEXT NOT NULL,
  cost REAL NOT NULL,
  method TEXT CHECK ( method in ('standard', 'express', 'overnight') ) NOT NULL
);

CREATE TABLE shipping_details (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  shipping_profile_id INTEGER NOT NULL,
  track_number TEXT,

  FOREIGN KEY (shipping_profile_id) REFERENCES shipping_profiles(id)
);

CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    user_id INTEGER NOT NULL,
    shipping_detail_id INTEGER,
    created DATETIME NOT NULL,
    external_order_id TEXT NOT NULL,
    payment_link TEXT NOT NULL,
    is_paid BOOLEAN NOT NULL,
    order_status TEXT CHECK ( order_status in ('created', 'shipped', 'cancelled', 'delivered') ) NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (shipping_detail_id) REFERENCES shipping_details(id)
);

CREATE TABLE prints (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  picture_id INTEGER NOT NULL,
  paper_id INTEGER NOT NULL,
  order_id INTEGER CHECK ( cart_id is null AND order_id is not null ),
  cart_id INTEGER CHECK ( order_id is null AND cart_id is not null ),
  width REAL NOT NULL,
  height REAL NOT NULL,
  border_size REAL NOT NULL,
  crop_x INTEGER NOT NULL,
  crop_y INTEGER NOT NULL,
  cost REAL NOT NULL,
  quantity INTEGER NOT NULL,

  FOREIGN KEY (picture_id) REFERENCES pictures(id),
  FOREIGN KEY (paper_id) REFERENCES papers(id),
  FOREIGN KEY (order_id) REFERENCES orders(id)
);

CREATE TABLE carts (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id INTEGER NOT NULL UNIQUE,

  FOREIGN KEY (user_id) REFERENCES users(id)
)
