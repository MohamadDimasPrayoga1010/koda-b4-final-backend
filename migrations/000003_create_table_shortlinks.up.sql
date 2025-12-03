CREATE TABLE shortlinks (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),         
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL UNIQUE,     
    redirect_count INT DEFAULT 0,               
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

ALTER TABLE shortlinks
ADD COLUMN status VARCHAR(20) DEFAULT 'active';


CREATE TABLE shortlink_clicks (
    id SERIAL PRIMARY KEY,
    shortlink_id INT REFERENCES shortlinks(id) ON DELETE CASCADE,
    ip_address VARCHAR(50),
    user_agent TEXT,
    clicked_at TIMESTAMP DEFAULT now()
);
