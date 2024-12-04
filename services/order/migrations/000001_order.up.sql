CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DO $$ 
BEGIN 
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN 
    CREATE TYPE order_status AS ENUM ('pending', 'shipped', 'delivered'); 
  END IF; 
END $$;

CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  customer_id UUID NOT NULL,
  address VARCHAR(255) NOT NULL,
  email   VARCHAR(255) NOT NULL,
  total_price NUMERIC(10,2) NOT NULL,
  status order_status DEFAULT 'pending',
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),   
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id UUID NOT NULL,
  name VARCHAR(255) NOT NULL,
  quantity INT NOT NULL,
  image VARCHAR(255) NOT NULL,
  price NUMERIC(10,2) NOT NULL
);
