ALTER TABLE users
ADD COLUMN getcourse_stream_id BIGINT;

UPDATE users
SET getcourse_stream_id = getcourse_stream_ids[1]
WHERE cardinality(getcourse_stream_ids) > 0;

ALTER TABLE users
DROP COLUMN getcourse_stream_ids;