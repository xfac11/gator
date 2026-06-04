-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows(
    id,
    created_at,
    updated_at,
    user_id,
    feed_id
)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *
)
SELECT 
    inserted_feed_follow.*,
    feeds.name as feed_name,
    users.name as user_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users ON users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, feeds.name AS feed_name, users.name AS user_name
FROM feeds
JOIN feed_follows ON feeds.id = feed_follows.feed_id
JOIN users ON users.id = $1
WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;