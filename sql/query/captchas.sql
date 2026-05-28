-- name: GetOrCreateCaptcha :one
INSERT INTO captchas (captcha_id, session_id, grid_size, mine_count, grid, revealed, difficulty_level)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (captcha_id) DO NOTHING
RETURNING *;

-- name: CreateCaptcha :one
INSERT INTO captchas (session_id, grid_size, mine_count, grid, revealed, difficulty_level)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCaptchaBySessionID :one
SELECT *
FROM captchas
WHERE session_id = $1 AND created_at >= NOW() - INTERVAL '1 hours' AND (solved = false OR failed = false)
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateCaptcha :one
UPDATE captchas
SET revealed = $2, solved = $3, failed = $4, updated_at = NOW()
WHERE captcha_id = $1
RETURNING *;

-- name: PurgeOldCaptchas :exec
DELETE FROM captchas
WHERE created_at < NOW() - INTERVAL '24 hours';
