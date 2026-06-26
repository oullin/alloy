package tempo

import (
	"errors"

	periodpkg "github.com/oullin/alloy/api/tempo/period"
)

func (period Period) Values() ([]Time, error) {
	values := make([]Time, 0)
	current := period.Start
	forward := period.End.SameOrAfter(period.Start)

	for {
		include := current.Before(period.End)

		if forward {
			include = current.Before(period.End)

			if period.IncludeEnd {
				include = current.SameOrBefore(period.End)
			}
		} else {
			include = current.After(period.End)

			if period.IncludeEnd {
				include = current.SameOrAfter(period.End)
			}
		}

		if !include {
			break
		}

		values = append(values, current)
		next := current.AddDuration(period.Step)

		if next.Same(current) {
			return nil, errors.New("tempo period step must advance the period")
		}

		if (forward && next.Before(current)) || (!forward && next.After(current)) {
			return nil, errors.New("tempo period step must advance toward the end")
		}

		current = next
	}

	return values, nil
}

func (period Period) First() (Time, bool, error) {
	values, err := period.Values()

	if err != nil {
		return Time{}, false, err
	}

	if len(values) == 0 {
		return Time{}, false, nil
	}

	return values[0], true, nil
}

func (period Period) Last() (Time, bool, error) {
	values, err := period.Values()

	if err != nil {
		return Time{}, false, err
	}

	if len(values) == 0 {
		return Time{}, false, nil
	}

	return values[len(values)-1], true, nil
}

func (period Period) Count() (int, error) {
	values, err := period.Values()

	if err != nil {
		return 0, err
	}

	return len(values), nil
}

func (period Period) IsEmpty() (bool, error) {
	count, err := period.Count()

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (period Period) Contains(input Time) bool {
	return periodpkg.Bounds{
		StartMs:    period.Start.TimestampMs(),
		EndMs:      period.End.TimestampMs(),
		IncludeEnd: period.IncludeEnd,
	}.Contains(input.TimestampMs())
}

func (period Period) Filter(predicate func(Time, int) bool) ([]Time, error) {
	values, err := period.Values()

	if err != nil {
		return nil, err
	}

	filtered := make([]Time, 0, len(values))

	for index, value := range values {
		if predicate(value, index) {
			filtered = append(filtered, value)
		}
	}

	return filtered, nil
}

func (period Period) Map(mapper func(Time, int) Time) ([]Time, error) {
	values, err := period.Values()

	if err != nil {
		return nil, err
	}

	mapped := make([]Time, 0, len(values))

	for index, value := range values {
		mapped = append(mapped, mapper(value, index))
	}

	return mapped, nil
}

func (period Period) Every(step Duration) Period {
	period.Step = step

	return period
}

func (period Period) ToDuration() Duration {
	return period.Start.IntervalUntil(period.End).ToDuration()
}
