package util

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	// Pre-compiled Unicode transformer for removing diacritics
	diacriticRemover = transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)

	// Pre-compiled character-to-character replacements
	charReplacements = map[rune]rune{
		// German
		'ß': 's',

		// Icelandic/Old English
		'ð': 'd',
		'þ': 't',

		// Polish
		'ł': 'l',

		// Scandinavian
		'ø': 'o',
		'å': 'a',

		// Slavic
		'č': 'c',
		'ć': 'c',
		'ž': 'z',
		'š': 's',
		'đ': 'd',
		'ř': 'r',
		'ň': 'n',
		'ť': 't',
		'ď': 'd',
		'ľ': 'l',
		'ĺ': 'l',
		'ŕ': 'r',

		// Romanian
		'ă': 'a',
		'â': 'a',
		'î': 'i',
		'ș': 's',
		'ț': 't',

		// Turkish
		'ğ': 'g',
		'ı': 'i',
		'ş': 's',

		// Hungarian
		'ő': 'o',
		'ű': 'u',

		// Spanish
		'ñ': 'n',

		// French/Portuguese
		'ç': 'c',

		// Currency symbols
		'€': 'e',
		'£': 'l',
		'$': 's',

		// Greek letters
		'α': 'a',
		'β': 'b',
		'γ': 'g',
		'δ': 'd',
		'ε': 'e',
		'ζ': 'z',
		'η': 'e',
		'θ': 't',
		'ι': 'i',
		'κ': 'k',
		'λ': 'l',
		'μ': 'm',
		'ν': 'n',
		'ξ': 'x',
		'ο': 'o',
		'π': 'p',
		'ρ': 'r',
		'σ': 's',
		'τ': 't',
		'υ': 'u',
		'φ': 'f',
		'χ': 'c',
		'ψ': 'p',
		'ω': 'o',
	}

	// Pre-compiled string replacer for multi-character replacements
	textReplacer = strings.NewReplacer(
		// German
		"ß", "ss",

		// Ligatures
		"æ", "ae",
		"œ", "oe",
		"ĳ", "ij",

		// Icelandic
		"þ", "th",

		// Typography ligatures
		"ﬀ", "ff",
		"ﬁ", "fi",
		"ﬂ", "fl",
		"ﬃ", "ffi",
		"ﬄ", "ffl",

		// Remove quotes
		"'", "",
		`"`, "",
		"«", "",
		"»", "",
		"‚", "",
		"„", "",

		// Normalize dashes
		"–", "-",
		"—", "-",
		"―", "-",

		// Normalize spaces
		" ", " ",
		" ", " ",
		" ", " ",

		// Other characters
		"…", "...",
	)
)

// NormalizeTextComprehensive removes diacritics and converts special characters
// to their ASCII equivalents for search purposes
func NormalizeText(s string) string {
	if s == "" {
		return ""
	}

	// Convert to lowercase first
	s = strings.ToLower(s)

	// Apply Unicode normalization to remove diacritics
	result, _, err := transform.String(diacriticRemover, s)
	if err != nil {
		result = s // fallback
	}

	// Apply character replacements
	runes := []rune(result)
	for i, r := range runes {
		if replacement, exists := charReplacements[r]; exists {
			runes[i] = replacement
		}
	}
	result = string(runes)

	// Apply multi-character replacements
	result = textReplacer.Replace(result)

	// Clean up spaces and trim
	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}
