-- Update salary ranges: $100k through $300k+

-- Delete existing salary ranges
DELETE FROM "RevenueRange" WHERE category = 'salary';

-- Insert new salary brackets ($20k increments, starting at $100k)
INSERT INTO "RevenueRange" (category, label, min_value, max_value, display_order) VALUES
('salary', '$100k - $120k', 100000, 120000, 0),
('salary', '$120k - $140k', 120000, 140000, 1),
('salary', '$140k - $160k', 140000, 160000, 2),
('salary', '$160k - $180k', 160000, 180000, 3),
('salary', '$180k - $200k', 180000, 200000, 4),
('salary', '$200k - $220k', 200000, 220000, 5),
('salary', '$220k - $240k', 220000, 240000, 6),
('salary', '$240k - $260k', 240000, 260000, 7),
('salary', '$260k - $280k', 260000, 280000, 8),
('salary', '$280k - $300k', 280000, 300000, 9),
('salary', '$300k+', 300000, NULL, 10);
