-- Enable UUID generator (Postgres 13+: pgcrypto has gen_random_uuid)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Bid history table
CREATE TABLE IF NOT EXISTS bids (
                                    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id  text        NOT NULL,
    bidder_id   text        NOT NULL,
    amount      numeric(18,2) NOT NULL CHECK (amount > 0),
    seq         bigserial   NOT NULL,           -- monotonic sequence (global)
    at          timestamptz NOT NULL DEFAULT now()
    );

-- Speed up "latest bid by auction"
CREATE INDEX IF NOT EXISTS bids_auction_seq_desc_idx
    ON bids (auction_id, seq DESC);