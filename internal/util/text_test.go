package util

import "testing"

func TestNormalizeText(t *testing.T) {
	t.Parallel()
	// Arrange
	tests := []struct {
		input    string
		expected string
	}{
		{"café", "cafe"},
		{"Cafe\u0301", "cafe"},
		{"façade", "facade"},
		{"naïve", "naive"},
		{"jalapeño", "jalapeno"},
		{"résumé", "resume"},
		{"crème brûlée", "creme brulee"},
		{"straße", "strasse"},
		{"façades", "facades"},
		{"français", "francais"},
		{"über", "uber"},
		{"þing", "thing"},
		// Ligatures
		{"encyclopædia", "encyclopaedia"},
		{"cœur", "coeur"},
		{"ĳssel", "ijssel"},
		// Spaces and zero-width characters
		{"A\u00A0B", "a b"},
		{"A\u200B\u200D\u200CB", "ab"},
		// Dashes and minus signs
		{"a–b—c−d‑e", "a-b-c-d-e"},
		// Quotes
		{"\"l'été\"", "lete"},
		// Greek characters (policy to be verified)
		{"θεός", "teos"}, // Note: Adjust based on intended policy (e.g., θ→"t" vs "th")
		{"χάος", "caos"}, // Note: Adjust based on intended policy (e.g., χ→"c" vs "ch")
		// Additional test cases
		{"Banjalučko", "banjalucko"},
		{"Müller", "muller"},
		{"José María", "jose maria"},
		{"François", "francois"},
		{"Škoda", "skoda"},
		{"Åse Bjørn", "ase bjorn"},
		{"Döner Kebab", "doner kebab"},
		{"Ürümqi", "urumqi"},
		{"Włochy", "wlochy"},
		{"Zürich", "zurich"},
		{"Ñoño", "nono"},
		{"Café", "cafe"},
		{"Naïve", "naive"},
		{"Résumé", "resume"},
		{"Piña Colada", "pina colada"},
		{"Björk", "bjork"},
		{"Čeština", "cestina"},
		{"İstanbul", "istanbul"},
		{"Kraków", "krakow"},
		{"Señorita", "senorita"},
		{"Château", "chateau"},
		{"Mädchen", "madchen"},
		{"Cañón", "canon"},
		{"Budějovice", "budejovice"},
		{"Łódź", "lodz"},
		{"“l’été”", "lete"}, // covers U+201C/U+201D and U+2019
		{"l’ete", "lete"},     // U+2019
		{"A\uFEFFB", "ab"},
		{"A\u2060B", "ab"},
		{"a‒b–c—d−e‑f", "a-b-c-d-e-f"}, // U+2012, U+2013, U+2014, U+2212, U+2011
		{"", ""},                       // empty input
		{" ", ""},                      // spaces collapse/trim
		{" A B C ", "a b c"},           // multi-space normalization
		{"B-52", "b-52"},               // ensure digits/punctuation stay as intended
		{"coöperate", "cooperate"},     // diaeresis
		{"Smørrebrød", "smorrebrod"},   // slashed o
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			// Act
			res := NormalizeText(tc.input)

			// Assert
			if res != tc.expected {
				t.Errorf("NormalizeText(%q) = %q; want %q", tc.input, res, tc.expected)
			}
		})
	}
}
