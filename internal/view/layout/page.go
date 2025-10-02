package layout

import (
	"github.com/a-h/templ"
)

type Page struct {
	Title     string
	URL       string
	Component templ.Component
}
