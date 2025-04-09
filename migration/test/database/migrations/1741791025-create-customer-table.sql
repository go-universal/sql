-- { up: table }
CREATE TABLE IF NOT EXISTS customers (id SERIAL, name text);

-- { down: table }
DROP TABLE IF EXISTS customers CASCADE;

-- { up: index }

-- { down: index }

