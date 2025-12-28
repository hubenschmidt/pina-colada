-- Add aggregate columns to recording sessions for data analysis
ALTER TABLE "Agent_Recording_Session"
ADD COLUMN IF NOT EXISTS metric_count INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_tokens INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_input_tokens INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_output_tokens INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_cost_usd DECIMAL(10, 6) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_duration_ms INT NOT NULL DEFAULT 0;

-- Backfill existing sessions with calculated aggregates
UPDATE "Agent_Recording_Session" s
SET
    metric_count = agg.cnt,
    total_tokens = agg.tokens,
    total_input_tokens = agg.input_tokens,
    total_output_tokens = agg.output_tokens,
    total_cost_usd = COALESCE(agg.cost, 0),
    total_duration_ms = agg.duration
FROM (
    SELECT
        session_id,
        COUNT(*) as cnt,
        SUM(total_tokens) as tokens,
        SUM(input_tokens) as input_tokens,
        SUM(output_tokens) as output_tokens,
        SUM(estimated_cost_usd) as cost,
        SUM(duration_ms) as duration
    FROM "Agent_Metric"
    GROUP BY session_id
) agg
WHERE s.id = agg.session_id;
