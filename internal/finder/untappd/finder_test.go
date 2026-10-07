package untappd

import "testing"

func TestBreweryPath(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		pageURL string
		id      int
		want    string
	}{
		"page name": {
			pageURL: "/Guinness",
			id:      49,
			want:    "/w/Guinness/49",
		},
		"path already contains ID": {
			pageURL: "/w/rewe/281970",
			id:      281970,
			want:    "/w/rewe/281970",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := breweryPath(test.pageURL, test.id)
			if got != test.want {
				t.Fatalf("breweryPath() = %q, want %q", got, test.want)
			}
		})
	}
}
