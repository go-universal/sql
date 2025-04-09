-- { up: table }
CREATE TABLE IF NOT EXISTS users (
    `id` SERIAL,
    `name` TEXT,
    `age` INTEGER
);

-- { down: table }
DROP TABLE IF EXISTS users CASCADE;

-- { up: index }
-- some code


-- { down: index }


