CREATE TABLE IF NOT EXISTS merch (
    id SERIAL PRIMARY KEY,
    merch_name VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL
);

CREATE TABLE IF NOT EXISTS employees (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    balance INT NOT NULL,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    sender_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    receiver_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    amount INT NOT NULL 
);


CREATE TABLE IF NOT EXISTS purchases (
    employee_id INT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    merch_id INT NOT NULL REFERENCES merch(id) ON DELETE CASCADE,
    count INT NOT NULL,
    PRIMARY KEY(employee_id, merch_id)
);

INSERT INTO merch (merch_name, price) VALUES
('t-shirt', 80),
('cup', 20),
('book', 50),
('pen', 10),
('powerbank', 200),
('hoody', 300),
('umbrella', 200),
('socks', 10),
('wallet', 50),
('pink-hoody', 500)
ON CONFLICT (merch_name) DO NOTHING;
