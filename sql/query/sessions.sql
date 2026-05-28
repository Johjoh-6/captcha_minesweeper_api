-- name: CreateSession :one
INSERT INTO sessions (ip_address, user_agent)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrCreateSession :one
INSERT INTO sessions (session_id, ip_address, user_agent)
VALUES ($1, $2, $3)
ON CONFLICT (session_id)
DO UPDATE SET updated_at = NOW()
RETURNING *;

-- name: UpdateSession :one
UPDATE sessions
SET
    updated_at = NOW(),
    request_count = request_count + 1,
    avg_time_between_requests =
        CASE
            WHEN request_count = 0 THEN GREATEST(EXTRACT(EPOCH FROM (NOW() - updated_at)) * 1000, 0)
            ELSE (
                COALESCE(avg_time_between_requests, 0) * request_count +
                GREATEST(EXTRACT(EPOCH FROM (NOW() - updated_at)) * 1000, 0)
            ) / (request_count + 1)
        END,
    is_bot =
        CASE
            WHEN is_bot THEN TRUE
            WHEN request_count >= 1 AND (
                    (avg_time_between_requests IS NOT NULL AND avg_time_between_requests > 0 AND avg_time_between_requests < 100) OR
                    (request_count >= 30 AND EXTRACT(EPOCH FROM (NOW() - updated_at)) * 1000 < 60000)
            )
            THEN TRUE
            ELSE is_bot
        END
WHERE session_id = $1
RETURNING *;
