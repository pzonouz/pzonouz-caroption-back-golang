-- =====================================================
-- Extensions
-- =====================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- Tables
-- =====================================================
CREATE TABLE IF NOT EXISTS categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE,
    description text,
    image_id uuid,
    priority varchar,
    parent_id uuid,
    show boolean,
    slug text,
    generator boolean,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS entities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE,
    description text,
    image_id uuid,
    price text,
    priority varchar,
    parent_id uuid,
    show boolean,
    keywords varchar [],
    slug text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS brands (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE,
    description text,
    info text,
    price text,
    image_id uuid,
    entity_id uuid,
    count text,
    Entity_id uuid,
    brand_id uuid,
    entitySlug text,
    generated boolean DEFAULT FALSE,
    generatable boolean DEFAULT FALSE,
    keywords varchar [],
    show boolean,
    position varchar,
    rank double precision,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS articles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    description text,
    image_id uuid,
    slug text,
    show_in_products boolean DEFAULT FALSE,
    category_id uuid,
    keywords varchar [],
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar,
    image_url text,
    product_id uuid,
    Entity_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameter_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    Entity_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    description text,
    type varchar,
    parameter_group_id uuid,
    selectables varchar [],
    priority varchar DEFAULT '10000000',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS product_parameter_values (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid,
    parameter_id uuid,
    text_value varchar,
    bool_value boolean,
    selectable_value varchar,
    created_at timestamptz DEFAULT now(),
    UNIQUE (parameter_id, product_id)
);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar UNIQUE,
    password text,
    token text,
    token_expires timestamptz,
    is_admin boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- =====================================================
-- Foreign keys
-- =====================================================
ALTER TABLE
    IF EXISTS product_parameter_values
ADD
    CONSTRAINT fk_ppv_parameter FOREIGN KEY (parameter_id) REFERENCES parameters (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS product_parameter_values
ADD
    CONSTRAINT fk_ppv_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS parameter_groups
ADD
    CONSTRAINT fk_parameter_groups_Entity FOREIGN KEY (Entity_id) REFERENCES categories (id);

ALTER TABLE
    IF EXISTS parameters
ADD
    CONSTRAINT fk_parameters_group FOREIGN KEY (parameter_group_id) REFERENCES parameter_groups (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS images
ADD
    CONSTRAINT fk_images_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE
SET
    NULL ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS images
ADD
    CONSTRAINT fk_images_Entity FOREIGN KEY (Entity_id) REFERENCES categories (id) ON DELETE
SET
    NULL ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS products
ADD
    CONSTRAINT fk_products_brand FOREIGN KEY (brand_id) REFERENCES brands (id);

ALTER TABLE
    IF EXISTS products
ADD
    CONSTRAINT fk_products_entity FOREIGN KEY (entity_id) REFERENCES entities (id);

ALTER TABLE
    IF EXISTS articles
ADD
    CONSTRAINT fk_articles_Entity FOREIGN KEY (Entity_id) REFERENCES categories (id);

ALTER TABLE
    IF EXISTS categories
ADD
    CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories (id);

ALTER TABLE
    IF EXISTS entities
ADD
    CONSTRAINT fk_entities_parent FOREIGN KEY (parent_id) REFERENCES entities (id);

ALTER TABLE
    IF EXISTS products
ADD
    CONSTRAINT fk_products_Entity FOREIGN KEY (Entity_id) REFERENCES categories (id);

-- =====================================================
-- Trigger: auto-update updated_at column
-- =====================================================
CREATE
OR REPLACE FUNCTION update_updated_at_column() RETURNS trigger AS $ $ BEGIN NEW.updated_at = now();

RETURN NEW;

END;

$ $ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_updated_at_products ON products;

CREATE TRIGGER set_updated_at_products BEFORE
UPDATE
    ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS set_updated_at_categories ON categories;

CREATE TRIGGER set_updated_at_categories BEFORE
UPDATE
    ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS set_updated_at_articles ON articles;

CREATE TRIGGER set_updated_at_articles BEFORE
UPDATE
    ON articles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS set_updated_at_users ON users;

CREATE TRIGGER set_updated_at_users BEFORE
UPDATE
    ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- Trigger: normalize Persian digits
-- =====================================================
CREATE
OR REPLACE FUNCTION normalize_persian_digits() RETURNS trigger AS $ $ DECLARE joined text;

normalized text;

BEGIN -- Fix Persian digits in priority
IF NEW.priority IS NOT NULL THEN NEW.priority := translate(NEW.priority, '۰۱۲۳۴۵۶۷۸۹', '0123456789');

END IF;

-- Fix Persian digits in selectables array
IF NEW.selectables IS NOT NULL THEN joined := array_to_string(NEW.selectables, '|');

normalized := translate(joined, '۰۱۲۳۴۵۶۷۸۹', '0123456789');

NEW.selectables := string_to_array(normalized, '|');

END IF;

RETURN NEW;

END;

$ $ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS normalize_persian_digits_trigger ON parameters;

CREATE TRIGGER normalize_persian_digits_trigger BEFORE
INSERT
    OR
UPDATE
    ON parameters FOR EACH ROW EXECUTE FUNCTION normalize_persian_digits();

-- =====================================================
-- Trigger: normalize Persian digits in product price/count
-- =====================================================
CREATE
OR REPLACE FUNCTION normalize_persian_digits_in_products() RETURNS trigger AS $ $ BEGIN IF NEW.price IS NOT NULL THEN NEW.price := translate(NEW.price, '۰۱۲۳۴۵۶۷۸۹', '0123456789');

END IF;

IF NEW.count IS NOT NULL THEN NEW.count := translate(NEW.count, '۰۱۲۳۴۵۶۷۸۹', '0123456789');

END IF;

RETURN NEW;

END;

$ $ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS normalize_persian_digits_in_products_trigger ON products;

CREATE TRIGGER normalize_persian_digits_in_products_trigger BEFORE
INSERT
    OR
UPDATE
    ON products FOR EACH ROW EXECUTE FUNCTION normalize_persian_digits_in_products();