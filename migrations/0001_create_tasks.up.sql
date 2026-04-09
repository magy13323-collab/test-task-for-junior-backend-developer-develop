CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	scheduled_at TIMESTAMPTZ NOT NULL,
	recurrence_type TEXT,
	recurrence_every_days INTEGER,
	recurrence_day_of_month INTEGER,
	recurrence_dates JSONB NOT NULL DEFAULT '[]'::jsonb,
	recurrence_day_parity TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_at ON tasks (scheduled_at);
