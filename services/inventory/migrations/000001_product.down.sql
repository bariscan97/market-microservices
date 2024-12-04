DROP TRIGGER IF EXISTS product_update_trigger ON products;

DROP FUNCTION IF EXISTS notify_product_update;
DROP FUNCTION IF EXISTS notify_product_delete;
DROP FUNCTION IF EXISTS notify_product_insert;

DROP TABLE IF EXISTS products;

DROP EXTENSION IF EXISTS "uuid-ossp";

