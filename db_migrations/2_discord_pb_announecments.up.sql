CREATE TABLE IF NOT EXISTS discord_pb_announecments (
  week DATE NOT NULL,
  message_id VARCHAR NOT NULL,
  part INT NOT NULL,
  total_parts INT NOT NULL,
  PRIMARY KEY (week, part)
);