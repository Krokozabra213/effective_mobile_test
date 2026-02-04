-- name: CreateSubscription :one
INSERT INTO subscriptions (
    service_name,
    price,
    user_id,
    start_date,
    end_date
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT *
FROM subscriptions
WHERE id = $1;

-- name: ListSubscriptions :many
SELECT *
FROM subscriptions
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListSubscriptionsByUserID :many
SELECT *
FROM subscriptions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteSubscription :execrows
DELETE FROM subscriptions
WHERE id = $1;
