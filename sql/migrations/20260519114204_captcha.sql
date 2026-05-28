-- +goose Up
CREATE TABLE captchas (
    captcha_id UUID PRIMARY KEY DEFAULT uuidv4(),
    session_id UUID NOT NULL REFERENCES sessions(session_id) ON DELETE CASCADE, -- Link to session
    grid_size INT NOT NULL,                -- e.g., 5, 8, 10
    mine_count INT NOT NULL,               -- e.g., 3, 10, 20
    grid INT[][] NOT NULL,                 -- 2D array: -1 = mine, 0-8 = adjacent mines
    revealed BOOLEAN[][] NOT NULL,         -- 2D array: revealed cells
    solved BOOLEAN DEFAULT FALSE,          -- True if user solved the CAPTCHA
    failed BOOLEAN DEFAULT FALSE,          -- True if user hit a mine (game over)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    difficulty_level INT DEFAULT 1        -- 1=noob, 2=very_easy, 3=easy, 4=medium, ...
);

-- Indexes
CREATE INDEX idx_captchas_session_id ON captchas(session_id);

-- +goose Down
DROP TABLE captchas;
