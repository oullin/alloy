package period

type Bounds struct {
	StartMs    int64
	EndMs      int64
	IncludeEnd bool
}

func (bounds Bounds) Forward() bool {
	return bounds.EndMs >= bounds.StartMs
}

func (bounds Bounds) Contains(inputMs int64) bool {
	if bounds.Forward() {
		if bounds.IncludeEnd {
			return inputMs >= bounds.StartMs && inputMs <= bounds.EndMs
		}

		return inputMs >= bounds.StartMs && inputMs < bounds.EndMs
	}

	if bounds.IncludeEnd {
		return inputMs <= bounds.StartMs && inputMs >= bounds.EndMs
	}

	return inputMs <= bounds.StartMs && inputMs > bounds.EndMs
}
