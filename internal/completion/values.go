package completion

// rawValues is a list of completion candidates.
type rawValues []Candidate

// Filter filters values.
func (c rawValues) Filter(values ...string) rawValues {
	toremove := make(map[string]bool)
	for _, v := range values {
		toremove[v] = true
	}

	filtered := make([]Candidate, 0)

	for _, rawValue := range c {
		if _, ok := toremove[rawValue.Value]; !ok {
			filtered = append(filtered, rawValue)
		}
	}

	return filtered
}

func (c *Values) merge(other Values) {
	if other.Usage != "" {
		c.Usage = other.Usage
	}

	c.NoSpace.Merge(other.NoSpace)
	c.Messages.Merge(other.Messages)

	for tag := range other.ListLong {
		if _, found := c.ListLong[tag]; !found {
			c.ListLong[tag] = true
		}
	}
}

func (c rawValues) eachTag(f func(tag string, values rawValues)) {
	tags := make([]string, 0)
	tagGroups := make(map[string]rawValues)

	for _, val := range c {
		if _, exists := tagGroups[val.Tag]; !exists {
			tagGroups[val.Tag] = make(rawValues, 0)

			tags = append(tags, val.Tag)
		}

		tagGroups[val.Tag] = append(tagGroups[val.Tag], val)
	}

	for _, tag := range tags {
		f(tag, tagGroups[tag])
	}
}
