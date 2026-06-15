package tempo

import (
	"fmt"
	"time"
)

func loadLocation(name string) (*time.Location, error) {
	if name == "" || name == "UTC" {
		return time.UTC, nil
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("invalid tempo time zone %q: %w", name, err)
	}

	return location, nil
}
