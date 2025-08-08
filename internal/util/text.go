package util

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizeText removes diacritics and converts special characters
// to their ASCII equivalents for search purposes
func NormalizeText(s string) string {
	// Step 1: Convert to lowercase
	s = strings.ToLower(s)

	// Step 2: Apply Unicode normalization (NFD) to decompose characters
	// This separates base characters from their diacritical marks
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, s)

	// Step 3: Handle specific characters that Unicode normalization might miss
	// or characters that need special treatment
	replacements := map[rune]rune{
		// German
		'ß': 's', // Will be handled by replacer below for "ss"

		// Icelandic/Old English
		'ð': 'd', 'þ': 't',
		'Ð': 'D', 'Þ': 'T',

		// Polish
		'ł': 'l', 'Ł': 'L',

		// Scandinavian
		'ø': 'o', 'Ø': 'O',
		'å': 'a', 'Å': 'A',

		// Slavic (in case Unicode normalization misses some)
		'č': 'c', 'Č': 'C',
		'ć': 'c', 'Ć': 'C',
		'ž': 'z', 'Ž': 'Z',
		'š': 's', 'Š': 'S',
		'đ': 'd', 'Đ': 'D',
		'ř': 'r', 'Ř': 'R',
		'ň': 'n', 'Ň': 'N',
		'ť': 't', 'Ť': 'T',
		'ď': 'd', 'Ď': 'D',
		'ľ': 'l', 'Ľ': 'L',
		'ĺ': 'l', 'Ĺ': 'L',
		'ŕ': 'r', 'Ŕ': 'R',

		// Romanian
		'ă': 'a', 'Ă': 'A',
		'â': 'a', 'Â': 'A',
		'î': 'i', 'Î': 'I',
		'ș': 's', 'Ș': 'S',
		'ț': 't', 'Ț': 'T',

		// Turkish
		'ğ': 'g', 'Ğ': 'G',
		'ı': 'i', 'İ': 'I',
		'ş': 's', 'Ş': 'S',

		// Hungarian
		'ő': 'o', 'Ő': 'O',
		'ű': 'u', 'Ű': 'U',

		// Other European
		'ñ': 'n', 'Ñ': 'N', // Spanish
		'ç': 'c', 'Ç': 'C', // French/Turkish/Portuguese

		// Currency and symbols that might appear in names
		'€': 'e', '£': 'l', '$': 's',
	}

	// Apply character-by-character replacements
	var normalized strings.Builder
	for _, r := range result {
		if replacement, exists := replacements[r]; exists {
			normalized.WriteRune(replacement)
		} else {
			normalized.WriteRune(r)
		}
	}

	result = normalized.String()

	// Step 4: Handle multi-character replacements and ligatures
	stringReplacer := strings.NewReplacer(
		// German
		"ß", "ss",

		// Ligatures
		"æ", "ae", "Æ", "AE",
		"œ", "oe", "Œ", "OE",
		"ĳ", "ij", "Ĳ", "IJ", // Dutch

		// Icelandic digraphs
		"þ", "th", "Þ", "TH",

		// Other multi-character cases
		"ﬀ", "ff", "ﬁ", "fi", "ﬂ", "fl", // Typography ligatures
		"ﬃ", "ffi", "ﬄ", "ffl",

		// Remove common punctuation that might interfere with search
		"'", "", "'", "", "'", "",
		`"`, "",
		"«", "", "»", "",
		"‚", "", "„", "",

		// Normalize various dashes and spaces
		"–", "-", "—", "-", "―", "-",
		" ", " ", " ", " ", // Various Unicode spaces to regular space

		// Remove or normalize other problematic characters
		"…", "...",
	)

	result = stringReplacer.Replace(result)

	// Step 5: Clean up multiple spaces and trim
	result = strings.Join(strings.Fields(result), " ")
	result = strings.TrimSpace(result)

	return result
}

// isMn checks if a rune is a nonspacing mark (diacritic)
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
