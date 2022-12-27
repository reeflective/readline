package readline

import (
	"fmt"
)

// Completion represents a completion candidate
type Completion struct {
	Value       string // Value is the value of the completion as actually inserted in the line
	Display     string // When display is not nil, this string is used to display the completion in the menu.
	Description string // A description to display next to the completion candidate.
	Style       string // An arbitrary string of color/text effects to use when displaying the completion.
	Tag         string // All completions with the same tag are grouped together and displayed under the tag heading.

	// A list of runes that are automatically trimmed when a space or a non-nil character is
	// inserted immediately after the completion. This is used for slash-autoremoval in path
	// completions, comma-separated completions, etc.
	noSpace suffixMatcher
}

// Completions holds all completions candidates and their associated data,
// including usage strings, messages, and suffix matchers for autoremoval.
// Some of those additional settings will apply to all contained candidates,
// except when these candidates have their own corresponding settings.
type Completions struct {
	values   rawValues
	messages messages
	noSpace  suffixMatcher
	usage    string
	listLong map[string]bool
	noSort   map[string]bool

	// Initially this will be set to the part of the current word
	// from the beginning of the word up to the position of the cursor;
	// it may be altered to give a common prefix for all matches.
	PREFIX string
}

// CompleteValues completes arbitrary keywords (values).
func CompleteValues(values ...string) Completions {
	vals := make([]Completion, 0, len(values))
	for _, val := range values {
		vals = append(vals, Completion{Value: val, Display: val, Description: ""})
	}
	return Completions{values: vals}
}

// CompleteStyledValues is like ActionValues but also accepts a style.
func CompleteStyledValues(values ...string) Completions {
	if length := len(values); length%2 != 0 {
		return Message("invalid amount of arguments [ActionStyledValues]: %v", length)
	}

	vals := make([]Completion, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		vals = append(vals, Completion{Value: values[i], Display: values[i], Description: "", Style: values[i+1]})
	}
	return Completions{values: vals}
}

// CompleteValuesDescribed completes arbitrary key (values) with an additional description (value, description pairs).
func CompleteValuesDescribed(values ...string) Completions {
	if length := len(values); length%2 != 0 {
		return Message("invalid amount of arguments [ActionValuesDescribed]: %v", length)
	}

	vals := make([]Completion, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		vals = append(vals, Completion{Value: values[i], Display: values[i], Description: values[i+1]})
	}
	return Completions{values: vals}
}

// CompleteStyledValuesDescribed is like ActionValues but also accepts a style.
func CompleteStyledValuesDescribed(values ...string) Completions {
	if length := len(values); length%3 != 0 {
		return Message("invalid amount of arguments [ActionStyledValuesDescribed]: %v", length)
	}

	vals := make([]Completion, 0, len(values)/3)
	for i := 0; i < len(values); i += 3 {
		vals = append(vals, Completion{Value: values[i], Display: values[i], Description: values[i+1], Style: values[i+2]})
	}
	return Completions{values: vals}
}

// CompleteRaw directly accepts a list of prepared Completion values.
func CompleteRaw(values []Completion) Completions {
	return Completions{values: rawValues(values)}
}

// Message displays a help messages in places where no completions can be generated.
func Message(msg string, args ...interface{}) Completions {
	c := Completions{}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	c.messages.Add(msg)
	return c
}

// Suppress suppresses specific error messages using regular expressions.
func (c Completions) Suppress(expr ...string) Completions {
	if err := c.messages.Suppress(expr...); err != nil {
		return Message(err.Error())
	}
	return c
}

// NoSpace disables space suffix for given characters (or all if none are given).
// These suffixes will be used for all completions that have not specified their
// own suffix-matching patterns.
// This is used for slash-autoremoval in path completions, comma-separated completions, etc.
func (c Completions) NoSpace(suffixes ...rune) Completions {
	if len(suffixes) == 0 {
		c.noSpace.Add('*')
	}
	c.noSpace.Add(suffixes...)
	return c
}

// Prefix adds a prefix to values (only the ones inserted, not the display values)
//
//	a := ActionValues("melon", "drop", "fall").Invoke(c)
//	b := a.Prefix("water") // ["watermelon", "waterdrop", "waterfall"] but display still ["melon", "drop", "fall"]
func (c Completions) Prefix(prefix string) Completions {
	for index, val := range c.values {
		c.values[index].Value = prefix + val.Value
	}
	return c
}

// Suffix adds a suffx to values (only the ones inserted, not the display values)
//
//	a := ActionValues("apple", "melon", "orange").Invoke(c)
//	b := a.Suffix("juice") // ["applejuice", "melonjuice", "orangejuice"] but display still ["apple", "melon", "orange"]
func (c Completions) Suffix(suffix string) Completions {
	for index, val := range c.values {
		c.values[index].Value = val.Value + suffix
	}
	return c
}

// Usage sets the usage.
func (c Completions) Usage(usage string, args ...interface{}) Completions {
	return c.UsageF(func() string {
		return fmt.Sprintf(usage, args...)
	})
}

// Usage sets the usage using a function.
func (c Completions) UsageF(f func() string) Completions {
	if usage := f(); usage != "" {
		c.usage = usage
	}
	return c
}

// Style sets the style, accepting cterm color codes, eg. 255, 30, etc.
//
//	ActionValues("yes").Style("35")
//	ActionValues("no").Style("255")
func (c Completions) Style(style string) Completions {
	return c.StyleF(func(s string) string {
		return style
	})
}

// Style sets the style using a reference
//
//	ActionValues("value").StyleR(&style.Value)
//	ActionValues("description").StyleR(&style.Value)
func (c Completions) StyleR(style *string) Completions {
	if style != nil {
		return c.Style(*style)
	}
	return c
}

// Style sets the style using a function
//
//	ActionValues("dir/", "test.txt").StyleF(myStyleFunc)
//	ActionValues("true", "false").StyleF(styleForKeyword)
func (c Completions) StyleF(f func(s string) string) Completions {
	for index, v := range c.values {
		c.values[index].Style = f(v.Value)
	}
	return c
}

// Tag sets the tag.
//
//	ActionValues("192.168.1.1", "127.0.0.1").Tag("interfaces").
func (c Completions) Tag(tag string) Completions {
	return c.TagF(func(value string) string {
		return tag
	})
}

// Tag sets the tag using a function.
//
//	ActionValues("192.168.1.1", "127.0.0.1").TagF(func(value string) string {
//		return "interfaces"
//	})
func (c Completions) TagF(f func(value string) string) Completions {
	for index, v := range c.values {
		c.values[index].Tag = f(v.Value)
	}
	return c
}

// DisplayList forces the completions to be list below each other as a list.
// A series of tags can be passed to restrict this to these tags. If empty,
// will be applied to all completions.
func (c Completions) DisplayList(tags ...string) Completions {
	if c.listLong == nil {
		c.listLong = make(map[string]bool)
	}
	if len(tags) == 0 {
		c.listLong["*"] = true
	}
	for _, tag := range tags {
		c.listLong[tag] = true
	}

	return c
}

// NoSort forces the completions not to sort the completions in alphabetical order.
// A series of tags can be passed to restrict this to these tags. If empty, will be
// applied to all completions.
func (c Completions) NoSort(tags ...string) Completions {
	if c.noSort == nil {
		c.noSort = make(map[string]bool)
	}
	if len(tags) == 0 {
		c.noSort["*"] = true
	}
	for _, tag := range tags {
		c.noSort[tag] = true
	}

	return c
}

// Filter filters given values (this should be done before any call to Prefix/Suffix as those alter the values being filtered)
//
//	a := ActionValues("A", "B", "C").Invoke(c)
//	b := a.Filter([]string{"B"}) // ["A", "C"]
func (c Completions) Filter(values []string) Completions {
	c.values = rawValues(c.values).Filter(values...)
	return c
}

// Merge merges Completions (existing values are overwritten)
//
//	a := ActionValues("A", "B").Invoke(c)
//	b := ActionValues("B", "C").Invoke(c)
//	c := a.Merge(b) // ["A", "B", "C"]
func (c Completions) Merge(others ...Completions) Completions {
	uniqueRawValues := make(map[string]Completion)
	for _, other := range append([]Completions{c}, others...) {
		for _, c := range other.values {
			uniqueRawValues[c.Value] = c
		}
	}

	for _, other := range others {
		c.merge(other)
	}

	rawValues := make([]Completion, 0, len(uniqueRawValues))
	for _, c := range uniqueRawValues {
		rawValues = append(rawValues, c)
	}

	c.values = rawValues
	return c
}

// rawValues is a list of completion candidates
type rawValues []Completion

// Filter filters values.
func (c rawValues) Filter(values ...string) rawValues {
	toremove := make(map[string]bool)
	for _, v := range values {
		toremove[v] = true
	}
	filtered := make([]Completion, 0)
	for _, rawValue := range c {
		if _, ok := toremove[rawValue.Value]; !ok {
			filtered = append(filtered, rawValue)
		}
	}
	return filtered
}

func (c *Completions) merge(other Completions) {
	if other.usage != "" {
		c.usage = other.usage
	}
	c.noSpace.Merge(other.noSpace)
	c.messages.Merge(other.messages)

	for tag := range other.listLong {
		if _, found := c.listLong[tag]; !found {
			c.listLong[tag] = true
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
