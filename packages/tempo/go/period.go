package tempo

import "errors"

func (period Period) Values() ([]Tempo, error) {
	values := make([]Tempo, 0)
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

func (period Period) First() (Tempo, bool, error) {
	values, err := period.Values()
	if err != nil {
		return Tempo{}, false, err
	}
	if len(values) == 0 {
		return Tempo{}, false, nil
	}

	return values[0], true, nil
}

func (period Period) Last() (Tempo, bool, error) {
	values, err := period.Values()
	if err != nil {
		return Tempo{}, false, err
	}
	if len(values) == 0 {
		return Tempo{}, false, nil
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

func (period Period) Contains(input Tempo) bool {
	forward := period.End.SameOrAfter(period.Start)
	afterStart := input.SameOrAfter(period.Start)
	if !forward {
		afterStart = input.SameOrBefore(period.Start)
	}

	beforeEnd := input.Before(period.End)
	if forward {
		if period.IncludeEnd {
			beforeEnd = input.SameOrBefore(period.End)
		}
	} else {
		beforeEnd = input.After(period.End)
		if period.IncludeEnd {
			beforeEnd = input.SameOrAfter(period.End)
		}
	}

	return afterStart && beforeEnd
}

func (period Period) Filter(predicate func(Tempo, int) bool) ([]Tempo, error) {
	values, err := period.Values()
	if err != nil {
		return nil, err
	}

	filtered := make([]Tempo, 0, len(values))
	for index, value := range values {
		if predicate(value, index) {
			filtered = append(filtered, value)
		}
	}

	return filtered, nil
}

func (period Period) Map(mapper func(Tempo, int) Tempo) ([]Tempo, error) {
	values, err := period.Values()
	if err != nil {
		return nil, err
	}

	mapped := make([]Tempo, 0, len(values))
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
