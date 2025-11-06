CREATE TABLE "categories" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" text UNIQUE,
    "description" text,
    "image_id" uuid,
    "prioirity" varchar,
    "parent_id" uuid,
    "show" boolean,
    "slug" text,
    "generator" boolean,
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "products" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" text UNIQUE,
    "description" text,
    "info" text,
    "price" text,
    "image_id" uuid,
    "count" text,
    "category_id" uuid,
    "brand_id" uuid,
    "slug" text,
    "generated" boolean,
    "generatable" boolean,
    "keywords" varchar[],
    "show" boolean,
    "Rank" float64,
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "articles" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar UNIQUE,
    "description" text,
    "image_id" uuid,
    "slug" text,
    "show_in_products" boolean DEFAULT (FALSE),
    "category_id" uuid,
    "keywords" varchar[],
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "images" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar,
    "image_url" text,
    "product_id" uuid,
    "category_id" uuid,
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "brands" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar UNIQUE,
    "description" text,
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "parameter_groups" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar UNIQUE,
    "category_id" uuid,
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "parameters" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar UNIQUE,
    "description" text,
    "type" varchar,
    "parameter_group_id" uuid,
    "selectables" varchar[],
    "prioirity" varchar DEFAULT ("10000000"),
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "product_parameter_values" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "product_id" uuid,
    "parameter_id" uuid,
    "text_value" varchar,
    "bool_value" boolean,
    "selectable_value" varchar,
    "created_at" timestamptz DEFAULT (now()),
    UNIQUE ("parameter_id", "product_id"),
    PRIMARY KEY ("id")
);

CREATE TABLE "users" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "email" varchar UNIQUE,
    "password" text,
    "is_admin" boolean DEFAULT (FALSE),
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

ALTER TABLE "product_parameter_values"
    ADD FOREIGN KEY ("parameter_id") REFERENCES "parameters" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "product_parameter_values"
    ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "parameter_groups"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "parameters"
    ADD FOREIGN KEY ("parameter_group_id") REFERENCES "parameter_groups" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE "images"
    ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "images"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "products"
    ADD FOREIGN KEY ("brand_id") REFERENCES "brands" ("id");

ALTER TABLE "articles"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "categories"
    ADD FOREIGN KEY ("parent_id") REFERENCES "categories" ("id");

ALTER TABLE "products"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

-- 1.Add the updated_at column to both tables
ALTER TABLE "products"
    ADD COLUMN IF NOT EXISTS "updated_at" timestamptz DEFAULT now();

ALTER TABLE "categories"
    ADD COLUMN IF NOT EXISTS "updated_at" timestamptz DEFAULT now();

ALTER TABLE "articles"
    ADD COLUMN IF NOT EXISTS "updated_at" timestamptz DEFAULT now();

-- 2.Create (or replace) the shared trigger function once
CREATE OR REPLACE FUNCTION update_updated_at_column ()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- 3.Create triggers for both tables
CREATE TRIGGER set_updated_at_products
    BEFORE UPDATE ON "products"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column ();

CREATE TRIGGER set_updated_at_categories
    BEFORE UPDATE ON "categories"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column ();

CREATE TRIGGER set_updated_at_articles
    BEFORE UPDATE ON "articles"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column ();

