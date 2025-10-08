CREATE TABLE "categories" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar UNIQUE,
    "description" text,
    "image_id" uuid,
    "prioirity" varchar,
    "parent_id" uuid,
    "created_at" timestamptz DEFAULT (now()),
    PRIMARY KEY ("id")
);

CREATE TABLE "products" (
    "id" uuid UNIQUE DEFAULT (gen_random_uuid ()),
    "name" varchar UNIQUE,
    "description" text,
    "info" varchar,
    "price" varchar,
    "image_id" uuid,
    "count" varchar,
    "category_id" uuid,
    "brand_id" uuid,
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
    PRIMARY KEY ("id")
);

ALTER TABLE "product_parameter_values"
    ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "product_parameter_values"
    ADD FOREIGN KEY ("parameter_id") REFERENCES "parameters" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "parameter_groups"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "products"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

ALTER TABLE "products"
    ADD FOREIGN KEY ("brand_id") REFERENCES "brands" ("id");

ALTER TABLE "images"
    ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "images"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "parameters"
    ADD FOREIGN KEY ("parameter_group_id") REFERENCES "parameter_groups" ("id");

ALTER TABLE "products"
    ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");

