UPDATE users SET target_chat_id = telegram_id WHERE target_chat_id IS NULL;
ALTER TABLE users ALTER COLUMN target_chat_id SET NOT NULL;