package layout

import (
	"github.com/a-h/templ"
)

type Page struct {
	// Context   echo.Context
	Title     string
	URL       string
	Component templ.Component
}

// func (p Page) Render() error {
// 	return p.Component.Render(p.Context.Request().Context(), p.Context.Response().Writer)
// }

// func NewPage(ctx echo.Context, title string) Page {
// 	page := Page{
// 		Title:   title,
// 		URL:     ctx.Request().URL.String(),
// 		Context: ctx,
// 	}
// 	return page
// }
