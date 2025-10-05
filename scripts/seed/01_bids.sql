INSERT INTO bids (auction_id, bidder_id, amount, at) VALUES
    ('a_seeded','u_1',100.00, now() - interval '2 minutes'),
    ('a_seeded','u_2',120.00, now() - interval '1 minutes');