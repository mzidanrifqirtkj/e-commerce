CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    parent_id BIGINT NULL,
    name varchar(100) NOT NULL,
    icon varchar(255) NOT NULL,
    status BOOLEAN DEFAULT TRUE,
    slug VARCHAR(120) UNIQUE NULL,
    description TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_categories_status ON categories(status);
CREATE INDEX idx_categories_slug ON categories(slug);
