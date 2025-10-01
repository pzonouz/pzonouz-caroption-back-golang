CREATE TABLE "categories" (
  "id" uuid UNIQUE DEFAULT (gen_random_uuid()),
  "name" varchar UNIQUE,
  "description" text,
  "prioirity" varchar,
  "parent_id" uuid,
  "created_at" timestamptz DEFAULT (now()),
  PRIMARY KEY ("id")
);

CREATE TABLE "products" (
  "id" uuid UNIQUE DEFAULT (gen_random_uuid()),
  "name" varchar UNIQUE,
  "description" text,
  "info" varchar,
  "price" varchar,
  "image_id" uuid UNIQUE,
  "count" varchar,
  "category_id" uuid,
  "brand_id" uuid,
  "created_at" timestamptz DEFAULT (now()),
  PRIMARY KEY ("id")
);

CREATE TABLE "images" (
  "id" uuid UNIQUE DEFAULT (gen_random_uuid()),
  "name" varchar,
  "image_url" text,
  "product_id" uuid,
  "category_id" uuid,
  "created_at" timestamptz DEFAULT (now()),
  PRIMARY KEY ("id")
);
CREATE TABLE "brands" (
  "id" uuid UNIQUE DEFAULT (gen_random_uuid()),
  "name" varchar UNIQUE,
  "description" text,
  "created_at" timestamptz DEFAULT (now()),
  PRIMARY KEY ("id")
);

ALTER TABLE "products" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");
ALTER TABLE "products" ADD FOREIGN KEY ("brand_id") REFERENCES "brands" ("id");

ALTER TABLE "images" ADD FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE "images" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE SET NULL ON UPDATE CASCADE ;


