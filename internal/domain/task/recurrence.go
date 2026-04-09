package task

import "time"

type RecurrenceType string

const (
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeDayParity     RecurrenceType = "day_parity"
)

type DayParity string

const (
	DayParityEven DayParity = "even"
	DayParityOdd  DayParity = "odd"
)

type Recurrence struct {
	Type       RecurrenceType `json:"type"`
	EveryDays  *int           `json:"every_days,omitempty"`
	DayOfMonth *int           `json:"day_of_month,omitempty"`
	Dates      []time.Time    `json:"dates,omitempty"`
	DayParity  DayParity      `json:"day_parity,omitempty"`
}

func (t RecurrenceType) Valid() bool {
	switch t {
	case RecurrenceTypeDaily, RecurrenceTypeMonthly, RecurrenceTypeSpecificDates, RecurrenceTypeDayParity:
		return true
	default:
		return false
	}
}

func (p DayParity) Valid() bool {
	switch p {
	case DayParityEven, DayParityOdd:
		return true
	default:
		return false
	}
}
