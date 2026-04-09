package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		ScheduledAt: normalized.ScheduledAt,
		Recurrence:  normalized.Recurrence,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		ScheduledAt: normalized.ScheduledAt,
		Recurrence:  normalized.Recurrence,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.ScheduledAt.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: scheduled_at is required", ErrInvalidInput)
	}
	input.ScheduledAt = input.ScheduledAt.UTC()

	normalizedRecurrence, err := validateRecurrence(input.Recurrence)
	if err != nil {
		return CreateInput{}, err
	}
	input.Recurrence = normalizedRecurrence

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.ScheduledAt.IsZero() {
		return UpdateInput{}, fmt.Errorf("%w: scheduled_at is required", ErrInvalidInput)
	}
	input.ScheduledAt = input.ScheduledAt.UTC()

	normalizedRecurrence, err := validateRecurrence(input.Recurrence)
	if err != nil {
		return UpdateInput{}, err
	}
	input.Recurrence = normalizedRecurrence

	return input, nil
}

func validateRecurrence(recurrence *taskdomain.Recurrence) (*taskdomain.Recurrence, error) {
	if recurrence == nil {
		return nil, nil
	}

	if !recurrence.Type.Valid() {
		return nil, fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	normalized := *recurrence

	switch normalized.Type {
	case taskdomain.RecurrenceTypeDaily:
		if normalized.EveryDays == nil {
			return nil, fmt.Errorf("%w: every_days is required for daily recurrence", ErrInvalidInput)
		}
		if *normalized.EveryDays <= 0 {
			return nil, fmt.Errorf("%w: every_days must be greater than 0 for daily recurrence", ErrInvalidInput)
		}
		if normalized.DayOfMonth != nil {
			return nil, fmt.Errorf("%w: day_of_month must be empty for daily recurrence", ErrInvalidInput)
		}
		if len(normalized.Dates) != 0 {
			return nil, fmt.Errorf("%w: dates must be empty for daily recurrence", ErrInvalidInput)
		}
		if normalized.DayParity != "" {
			return nil, fmt.Errorf("%w: day_parity must be empty for daily recurrence", ErrInvalidInput)
		}
	case taskdomain.RecurrenceTypeMonthly:
		if normalized.DayOfMonth == nil {
			return nil, fmt.Errorf("%w: day_of_month is required for monthly recurrence", ErrInvalidInput)
		}
		if *normalized.DayOfMonth < 1 || *normalized.DayOfMonth > 30 {
			return nil, fmt.Errorf("%w: day_of_month must be between 1 and 30 for monthly recurrence", ErrInvalidInput)
		}
		if normalized.EveryDays != nil {
			return nil, fmt.Errorf("%w: every_days must be empty for monthly recurrence", ErrInvalidInput)
		}
		if len(normalized.Dates) != 0 {
			return nil, fmt.Errorf("%w: dates must be empty for monthly recurrence", ErrInvalidInput)
		}
		if normalized.DayParity != "" {
			return nil, fmt.Errorf("%w: day_parity must be empty for monthly recurrence", ErrInvalidInput)
		}
	case taskdomain.RecurrenceTypeSpecificDates:
		if len(normalized.Dates) == 0 {
			return nil, fmt.Errorf("%w: dates must not be empty for specific_dates recurrence", ErrInvalidInput)
		}
		if normalized.EveryDays != nil {
			return nil, fmt.Errorf("%w: every_days must be empty for specific_dates recurrence", ErrInvalidInput)
		}
		if normalized.DayOfMonth != nil {
			return nil, fmt.Errorf("%w: day_of_month must be empty for specific_dates recurrence", ErrInvalidInput)
		}
		if normalized.DayParity != "" {
			return nil, fmt.Errorf("%w: day_parity must be empty for specific_dates recurrence", ErrInvalidInput)
		}

		unique := make([]time.Time, 0, len(normalized.Dates))
		seen := make(map[string]struct{}, len(normalized.Dates))
		for _, date := range normalized.Dates {
			if date.IsZero() {
				return nil, fmt.Errorf("%w: dates must not contain zero values", ErrInvalidInput)
			}
			utc := date.UTC()
			normalizedDate := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
			key := normalizedDate.Format("2006-01-02")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			unique = append(unique, normalizedDate)
		}
		if len(unique) == 0 {
			return nil, fmt.Errorf("%w: dates must not be empty after normalization", ErrInvalidInput)
		}
		normalized.Dates = unique
	case taskdomain.RecurrenceTypeDayParity:
		if !normalized.DayParity.Valid() {
			return nil, fmt.Errorf("%w: invalid day_parity", ErrInvalidInput)
		}
		if normalized.EveryDays != nil {
			return nil, fmt.Errorf("%w: every_days must be empty for day_parity recurrence", ErrInvalidInput)
		}
		if normalized.DayOfMonth != nil {
			return nil, fmt.Errorf("%w: day_of_month must be empty for day_parity recurrence", ErrInvalidInput)
		}
		if len(normalized.Dates) != 0 {
			return nil, fmt.Errorf("%w: dates must be empty for day_parity recurrence", ErrInvalidInput)
		}
	}

	return &normalized, nil
}
