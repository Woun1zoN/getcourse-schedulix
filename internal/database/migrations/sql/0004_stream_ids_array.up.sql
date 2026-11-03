ALTER TABLE users
ADD COLUMN getcourse_stream_ids BIGINT[] NOT NULL DEFAULT '{}';

UPDATE users
SET getcourse_stream_ids = ARRAY[getcourse_stream_id]
WHERE getcourse_stream_id IS NOT NULL;

ALTER TABLE users
DROP COLUMN getcourse_stream_id;