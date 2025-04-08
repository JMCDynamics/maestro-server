CREATE TYPE operating_system AS ENUM ('WINDOWS', 'LINUX');

CREATE TABLE nodes (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    vpn_address VARCHAR(50) NOT NULL UNIQUE,
    operating_system operating_system NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP NULL
);
