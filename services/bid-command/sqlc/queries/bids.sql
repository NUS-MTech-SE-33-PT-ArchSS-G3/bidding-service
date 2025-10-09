-- name: InsertBid :one
INSERT INTO bids (auction_id, bidder_id, amount, at)
VALUES ($1, $2, $3, $4)
    RETURNING id, seq;

-- name: LatestForUpdate :one
SELECT id, amount, seq, at
FROM bids
WHERE auction_id = $1
ORDER BY seq DESC
    LIMIT 1
    FOR UPDATE;