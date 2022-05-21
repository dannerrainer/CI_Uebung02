CREATE TABLE products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
);

CREATE TABLE ratings
(
        rating_id SERIAL PRIMARY KEY,
        product_id INTEGER REFERENCES products(id)
                            ON DELETE CASCADE
                            ON UPDATE CASCADE,
        rating INTEGER NOT NULL,
        info TEXT NULL
);

GRANT ALL PRIVILEGES ON TABLE ratings TO postgres;
GRANT ALL PRIVILEGES ON TABLE products TO postgres;

ALTER TABLE ratings OWNER TO postgres;
ALTER SEQUENCE ratings_rating_id_seq OWNER TO postgres;
