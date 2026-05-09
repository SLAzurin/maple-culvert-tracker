CREATE TABLE IF NOT EXISTS discord_monthly_improvements (
  month DATE PRIMARY KEY,
  message_id VARCHAR NOT NULL,
  channel_id VARCHAR NOT NULL,
  fluff_text TEXT NOT NULL
);
