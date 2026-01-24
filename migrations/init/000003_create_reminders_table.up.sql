-- Create reminders table
CREATE TABLE IF NOT EXISTS reminders (
    id          SERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    remind_at   TIMESTAMPTZ NOT NULL,
    is_sent     BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Index for fast user lookup
CREATE INDEX IF NOT EXISTS idx_reminders_user_id ON reminders(user_id);

-- Index for scheduler (future use)
CREATE INDEX IF NOT EXISTS idx_reminders_due ON reminders(remind_at) WHERE is_sent = FALSE;
