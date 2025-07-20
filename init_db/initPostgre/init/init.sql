CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE,
    pass TEXT,
    liked_listings UUID[] DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS listings (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    likes INT DEFAULT 0,
    description TEXT,
    address TEXT NOT NULL,
    price INT NOT NULL,
    author_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    image_url TEXT
);