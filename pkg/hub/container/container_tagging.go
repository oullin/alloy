package container

import "slices"

// Tag assigns one or more tags to the given abstracts.
func (c *App) Tag(abstracts []string, tags ...string) {
	c.mu.Lock()

	defer c.mu.Unlock()

	for _, tag := range tags {
		c.tags[tag] = append(c.tags[tag], abstracts...)
	}
}

// Tagged resolves all abstracts registered under the given tag.
func (c *App) Tagged(tag string) []any {
	c.mu.RLock()
	abstracts := slices.Clone(c.tags[tag])
	c.mu.RUnlock()

	var results []any

	for _, abstract := range abstracts {
		instance, err := c.Make(abstract)

		if err == nil {
			results = append(results, instance)
		}
	}

	return results
}
