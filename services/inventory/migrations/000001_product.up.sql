CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(255),
    description TEXT,
    image_url TEXT,
    price DECIMAL(10, 2) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    stock_quantity INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION notify_product_update()
RETURNS TRIGGER AS $$
DECLARE
    updated_fields JSONB := jsonb_build_object('id', NEW.id); 
    key TEXT;
    new_value TEXT;
    old_value TEXT;
BEGIN
    FOR key IN SELECT column_name FROM information_schema.columns WHERE table_name = 'products' LOOP
        EXECUTE format('SELECT ($1).%I::TEXT', key) INTO new_value USING NEW;
        EXECUTE format('SELECT ($1).%I::TEXT', key) INTO old_value USING OLD;

        IF new_value IS DISTINCT FROM old_value THEN
            updated_fields := jsonb_set(updated_fields, ARRAY[key], to_jsonb(new_value));
        END IF;
    END LOOP;

    PERFORM pg_notify('product_update', updated_fields::text);
    
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION notify_product_delete()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('product_delete', json_build_object('id', OLD.id)::text);
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION notify_product_insert()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('product_insert', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER product_insert_trigger
AFTER INSERT ON products
FOR EACH ROW
EXECUTE FUNCTION notify_product_insert();

CREATE TRIGGER product_update_trigger
AFTER UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION notify_product_update();

CREATE TRIGGER product_delete_trigger
AFTER DELETE ON products
FOR EACH ROW
EXECUTE FUNCTION notify_product_delete();