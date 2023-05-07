package strutil

import "github.com/reeflective/readline/inputrc"

// ConvertMeta recursively searches for metafied keys in a sequence,
// and replaces them with an esc prefix and their unmeta equivalent.
func ConvertMeta(keys []rune) string {
	if len(keys) == 0 {
		return string(keys)
	}

	converted := make([]rune, 0)

	for i := 0; i < len(keys); i++ {
		char := keys[i]

		if !inputrc.IsMeta(char) {
			converted = append(converted, char)
			continue
		}

		// Replace the key with esc prefix and add the demetafied key.
		converted = append(converted, inputrc.Esc)
		converted = append(converted, inputrc.Demeta(char))
	}

	return string(converted)
}
