DROP TABLE IF EXISTS publication, site, category;

CREATE TABLE category
(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE site
(
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    url TEXT NOT NULL,
    category_id INTEGER NOT NULL REFERENCES category(id)
);

CREATE TABLE publication
(
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    annotation TEXT,
    publication_time INTEGER NOT NULL,
    publication_url TEXT NOT NULL,
    site_id INTEGER NOT NULL REFERENCES site(id)
);