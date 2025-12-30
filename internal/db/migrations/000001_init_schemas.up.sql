-- =====================================================
-- Extensions
-- =====================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- Tables
-- =====================================================
CREATE TABLE IF NOT EXISTS categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
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

CREATE TABLE IF NOT EXISTS persons (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    first_name text,
    last_name text,
    address text,
    phone_number text UNIQUE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS entities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name text UNIQUE,
    description text,
    image_id uuid,
    price text,
    priority varchar,
    parent_id uuid,
    show boolean,
    keywords varchar[],
    slug text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS brands (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name varchar UNIQUE,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name text UNIQUE,
    description text,
    info text,
    price text,
    image_id uuid,
    entity_id uuid,
    count text,
    brand_id uuid,
    entitySlug text,
    generated boolean DEFAULT FALSE,
    generatable boolean DEFAULT FALSE,
    keywords varchar[],
    show boolean,
    position varchar,
    rank double precision,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS articles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name varchar UNIQUE,
    description text,
    image_id uuid,
    slug text,
    show_in_products boolean DEFAULT FALSE,
    category_id uuid,
    keywords varchar[],
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name varchar,
    image_url text,
    product_id uuid,
    Entity_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameter_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name varchar UNIQUE,
    Entity_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name varchar UNIQUE,
    description text,
    type varchar,
    parameter_group_id uuid,
    selectables varchar[],
    priority varchar DEFAULT '10000000',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS product_parameter_values (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    product_id uuid,
    parameter_id uuid,
    text_value varchar,
    bool_value boolean,
    selectable_value varchar,
    created_at timestamptz DEFAULT now(),
    UNIQUE (parameter_id, product_id)
);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    email varchar UNIQUE,
    password text,
    token text,
    token_expires timestamptz,
    is_admin boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TYPE invoice_type AS ENUM (
    'sell',
    'buy'
);

CREATE TABLE IF NOT EXISTS invoices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    person_id uuid NOT NULL,
    type invoice_type NOT NULL,
    number serial UNIQUE NOT NULL,
    total numeric(12, 2) NOT NULL DEFAULT 0,
    discount numeric(12, 2) NOT NULL DEFAULT 0,
    net_total numeric(12, 2) NOT NULL DEFAULT 0,
    notes text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoice_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    invoice_id uuid NOT NULL REFERENCES invoices (id) ON DELETE CASCADE,
    product_id uuid NULL,
    price numeric(12, 2) NOT NULL DEFAULT 0,
    discount numeric(12, 2) NOT NULL DEFAULT 0,
    count integer NOT NULL DEFAULT 1,
    description text,
    total numeric(12, 2) NOT NULL DEFAULT 0,
    net_total numeric(12, 2) NOT NULL DEFAULT 0,
    created_at timestamp with time zone DEFAULT now()
);

-- =====================================================
-- Foreign keys
-- =====================================================
ALTER TABLE IF EXISTS product_parameter_values
    ADD CONSTRAINT fk_ppv_parameter FOREIGN KEY (parameter_id) REFERENCES parameters (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE IF EXISTS product_parameter_values
    ADD CONSTRAINT fk_ppv_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE IF EXISTS parameter_groups
    ADD CONSTRAINT fk_parameter_groups_entity FOREIGN KEY (entity_id) REFERENCES entities (id);

ALTER TABLE IF EXISTS parameters
    ADD CONSTRAINT fk_parameters_group FOREIGN KEY (parameter_group_id) REFERENCES parameter_groups (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE IF EXISTS images
    ADD CONSTRAINT fk_images_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE IF EXISTS images
    ADD CONSTRAINT fk_images_entity FOREIGN KEY (entity_id) REFERENCES entities (id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE IF EXISTS products
    ADD CONSTRAINT fk_products_brand FOREIGN KEY (brand_id) REFERENCES brands (id);

ALTER TABLE IF EXISTS products
    ADD CONSTRAINT fk_products_entity FOREIGN KEY (entity_id) REFERENCES entities (id);

ALTER TABLE IF EXISTS categories
    ADD CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories (id);

ALTER TABLE IF EXISTS entities
    ADD CONSTRAINT fk_entities_parent FOREIGN KEY (parent_id) REFERENCES entities (id);

-- FK to persons table (adjust table/column name if different)
ALTER TABLE invoices
    ADD CONSTRAINT invoices_person_fk FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE RESTRICT;

-- Optional FK to products table (uncomment/adjust if product table exists)
ALTER TABLE invoice_items
    ADD CONSTRAINT invoice_items_product_fk FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT;

ALTER TABLE invoice_items
    ADD CONSTRAINT invoice_items_invoice_fk FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE CASCADE;


