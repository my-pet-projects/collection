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
	}

	for _, test := range tests {
		// Act
		res := NormalizeText(test.input)

		// Assert
		if res != test.expected {
			t.Errorf("NormalizeText(%q) = %q; want %q", test.input, res, test.expected)
		}
	}
}
