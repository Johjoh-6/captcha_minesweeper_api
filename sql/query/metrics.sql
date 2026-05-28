-- name: GetAppMetrics :one
SELECT
    (SELECT COUNT(*) FROM captchas) AS captchas_total,
    (SELECT COUNT(*) FROM captchas WHERE COALESCE(solved, false) = TRUE) AS captchas_solved,
    (SELECT COUNT(*) FROM captchas WHERE COALESCE(solved, false) = FALSE) AS captchas_unsolved,
    (SELECT COUNT(*) FROM captchas WHERE created_at >= NOW() - INTERVAL '1 hour') AS captchas_last_hour,
    (SELECT COUNT(*) FROM sessions) AS sessions_total,
    (SELECT COUNT(*) FROM sessions WHERE COALESCE(is_bot, false) = TRUE) AS sessions_bots,
    (SELECT COUNT(*) FROM sessions WHERE COALESCE(is_bot, false) = FALSE) AS sessions_non_bots;
