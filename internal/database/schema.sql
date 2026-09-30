CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'client',
    is_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    confirmation_code VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY,
    business_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DOUBLE PRECISION NOT NULL,
    category VARCHAR(100) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS carts (
    id INTEGER PRIMARY KEY,
    business_id VARCHAR(100) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cart_items (
    id INTEGER PRIMARY KEY,
    cart_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cart_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY,
    business_id VARCHAR(100) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    cart_id INT NOT NULL,
    total_ttc DOUBLE PRECISION NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'en attente',
    cancel_reason TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id INTEGER PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    unit_price DOUBLE PRECISION NOT NULL
);

INSERT INTO products (business_id, name, description, price, category, stock) 
SELECT 'PDT-LAP001', 'PC Portable Go', 'Ordinateur portable performant pour développeurs Go', 1199.99, 'Informatique', 10
WHERE NOT EXISTS (SELECT 1 FROM products WHERE business_id = 'PDT-LAP001');

INSERT INTO products (business_id, name, description, price, category, stock) 
SELECT 'PDT-CLV002', 'Clavier Mécanique RGB', 'Clavier mécanique rétroéclairé switchs blue', 89.90, 'Accessoires', 25
WHERE NOT EXISTS (SELECT 1 FROM products WHERE business_id = 'PDT-CLV002');

INSERT INTO products (business_id, name, description, price, category, stock) 
SELECT 'PDT-MOU003', 'Souris Ergonomique Sans Fil', 'Souris sans fil haute précision 4000 DPI', 49.99, 'Accessoires', 30
WHERE NOT EXISTS (SELECT 1 FROM products WHERE business_id = 'PDT-MOU003');

INSERT INTO products (business_id, name, description, price, category, stock) 
SELECT 'PDT-MON004', 'Écran 27 Pouces 4K', 'Moniteur 27 pouces IPS 4K UHD 144Hz', 349.50, 'Informatique', 15
WHERE NOT EXISTS (SELECT 1 FROM products WHERE business_id = 'PDT-MON004');
