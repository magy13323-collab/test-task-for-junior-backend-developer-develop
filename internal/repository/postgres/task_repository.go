package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (
			title, description, status, scheduled_at,
			recurrence_type, recurrence_every_days, recurrence_day_of_month, recurrence_dates, recurrence_day_parity,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id, title, description, status, scheduled_at,
			recurrence_type, recurrence_every_days, recurrence_day_of_month, recurrence_dates, recurrence_day_parity,
			created_at, updated_at
	`

	recurrenceType, everyDays, dayOfMonth, datesJSON, dayParity, err := recurrenceDBFields(task)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.ScheduledAt,
		recurrenceType,
		everyDays,
		dayOfMonth,
		datesJSON,
		dayParity,
		task.CreatedAt,
		task.UpdatedAt,
	)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT
			id, title, description, status, scheduled_at,
			recurrence_type, recurrence_every_days, recurrence_day_of_month, recurrence_dates, recurrence_day_parity,
			created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			scheduled_at = $4,
			recurrence_type = $5,
			recurrence_every_days = $6,
			recurrence_day_of_month = $7,
			recurrence_dates = $8,
			recurrence_day_parity = $9,
			updated_at = $10
		WHERE id = $11
		RETURNING
			id, title, description, status, scheduled_at,
			recurrence_type, recurrence_every_days, recurrence_day_of_month, recurrence_dates, recurrence_day_parity,
			created_at, updated_at
	`

	recurrenceType, everyDays, dayOfMonth, datesJSON, dayParity, err := recurrenceDBFields(task)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.ScheduledAt,
		recurrenceType,
		everyDays,
		dayOfMonth,
		datesJSON,
		dayParity,
		task.UpdatedAt,
		task.ID,
	)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT
			id, title, description, status, scheduled_at,
			recurrence_type, recurrence_every_days, recurrence_day_of_month, recurrence_dates, recurrence_day_parity,
			created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func recurrenceDBFields(task *taskdomain.Task) (recurrenceType *string, everyDays *int, dayOfMonth *int, datesJSON []byte, dayParity *string, err error) {
	datesJSON = []byte("[]")
	if task.Recurrence == nil {
		return nil, nil, nil, datesJSON, nil, nil
	}

	recurrenceTypeValue := string(task.Recurrence.Type)
	recurrenceType = &recurrenceTypeValue
	everyDays = task.Recurrence.EveryDays
	dayOfMonth = task.Recurrence.DayOfMonth

	if task.Recurrence.Dates != nil {
		datesJSON, err = json.Marshal(task.Recurrence.Dates)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	if task.Recurrence.DayParity != "" {
		dayParityValue := string(task.Recurrence.DayParity)
		dayParity = &dayParityValue
	}

	return recurrenceType, everyDays, dayOfMonth, datesJSON, dayParity, nil
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status         string
		recurrenceType *string
		everyDays      *int
		dayOfMonth     *int
		datesJSON      []byte
		dayParity      *string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.ScheduledAt,
		&recurrenceType,
		&everyDays,
		&dayOfMonth,
		&datesJSON,
		&dayParity,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)
	if recurrenceType != nil {
		recurrence := &taskdomain.Recurrence{
			Type:       taskdomain.RecurrenceType(*recurrenceType),
			EveryDays:  everyDays,
			DayOfMonth: dayOfMonth,
		}

		if len(datesJSON) > 0 {
			var dates []time.Time
			if err := json.Unmarshal(datesJSON, &dates); err != nil {
				return nil, err
			}
			recurrence.Dates = dates
		}

		if dayParity != nil {
			recurrence.DayParity = taskdomain.DayParity(*dayParity)
		}

		task.Recurrence = recurrence
	} else {
		task.Recurrence = nil
	}

	return &task, nil
}
