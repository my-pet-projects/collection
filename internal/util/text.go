package util

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Precompiled transforms/tables are read-only and reused for performance.
var (
	// Pre-compiled Unicode transformer for removing diacritics.
	diacriticRemover = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC) //nolint:gochecknoglobals

	// Pre-compiled character-to-character replacements.
	charReplacements = map[rune]rune{ //nolint:gochecknoglobals
		// Icelandic/Old English
		'ð': 'd',

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
		'ς': 's',
	}

	// Pre-compiled string replacer for multi-character replacements.
	textReplacer = strings.NewReplacer( //nolint:gochecknoglobals
		// Multi-character replacements (these take precedence)
		"ß", "ss", // German eszett
		"þ", "th", // Icelandic thorn

		// Ligatures
		"æ", "ae",
		"œ", "oe",
		"ĳ", "ij",

		// Typography ligatures
		"ﬀ", "ff",
		"ﬁ", "fi",
		"ﬂ", "fl",
		"ﬃ", "ffi",
		"ﬄ", "ffl",

		// Remove currency symbols
		"€", "",
		"£", "",
		"$", "",

		// Remove quotes and similar punctuation
		"'", "",
		"\"", "",
		"‘", "",
		"’", "",
		"“", "",
		"”", "",
		"«", "",
		"»", "",
		"‚", "",
		"„", "",

		// Normalize dashes to regular hyphen
		"–", "-",
		"—", "-",
		"―", "-",
		"‑", "-", // U+2011 No-Break hyphen
		"‒", "-", // U+2012 Figure Dash
		"−", "-", // U+2212 Minus sign
		"\u00AD", "", // Soft hyphen (discretionary)

		// Normalize various Unicode spaces to regular space
		"\u00A0", " ", // Non-breaking space
		"\u2002", " ", // En space
		"\u2003", " ", // Em space
		"\u2009", " ", // Thin space
		"\u200A", " ", // Hair space
		"\u202F", " ", // Narrow no-break space
		"\u200B", "", // Zero-width space
		"\u200C", "", // Zero-width non-joiner
		"\u200D", "", // Zero-width joiner
		"\u2060", "", // Word joiner

		// Other problematic characters
		"…", "...",
	)
)

// NormalizeText removes diacritics and converts special characters to ASCII-friendly equivalents for search purposes.
func NormalizeText(text string) string {
	if text == "" {
		return ""
	}

	// Convert to lowercase first
	text = strings.ToLower(text)

	// Apply Unicode normalization to remove diacritics or fallback
	result, _, err := transform.String(diacriticRemover, text)
	if err != nil {
		result = text
	}

	// Apply character replacements
	rr := []rune(result)
	for i, r := range rr {
		if replacement, exists := charReplacements[r]; exists {
			rr[i] = replacement
		}
	}
	result = string(rr)

	// Apply multi-character replacements
	result = textReplacer.Replace(result)

	// Clean up spaces and trim
	result = strings.Join(strings.Fields(result), " ")
	return strings.TrimSpace(result)
}
