// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.887
package workspace

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/my-pet-projects/collection/internal/view/component/navigation"
	"github.com/my-pet-projects/collection/internal/view/component/shared"
)

func adminLayout_todel(page Page, contents templ.Component) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<html class=\"light\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = headerComponent(page.Title).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = navigation.NavBar().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<div class=\"flex h-screen overflow-hidden pt-16\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = navigation.SideBar(navigation.SidebarData{URL: page.URL}).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<div id=\"main-content\" class=\"relative h-full w-full overflow-y-auto lg:ml-64\"><main>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if contents != nil {
			templ_7745c5c3_Err = contents.Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = shared.Notification().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "</div></main></div></div><script>\n\t\tconst sidebar = document.getElementById('sidebar');\n\t\tif (sidebar) {\n\t\t\tconst toggleSidebarMobile = (sidebar, sidebarBackdrop, toggleSidebarMobileHamburger, toggleSidebarMobileClose) => {\n\t\t\t\tsidebar.classList.toggle('hidden');\n\t\t\t\tsidebarBackdrop.classList.toggle('hidden');\n\t\t\t\ttoggleSidebarMobileHamburger.classList.toggle('hidden');\n\t\t\t\ttoggleSidebarMobileClose.classList.toggle('hidden');\n\t\t\t}\n\t\t\t\n\t\t\tconst toggleSidebarMobileEl = document.getElementById('toggleSidebarMobile');\n\t\t\tconst sidebarBackdrop = document.getElementById('sidebarBackdrop');\n\t\t\tconst toggleSidebarMobileHamburger = document.getElementById('toggleSidebarMobileHamburger');\n\t\t\tconst toggleSidebarMobileClose = document.getElementById('toggleSidebarMobileClose');\n\t\t\t\n\t\t\ttoggleSidebarMobileEl.addEventListener('click', () => {\n\t\t\t\ttoggleSidebarMobile(sidebar, sidebarBackdrop, toggleSidebarMobileHamburger, toggleSidebarMobileClose);\n\t\t\t});\n\t\t\t\n\t\t\tsidebarBackdrop.addEventListener('click', () => {\n\t\t\t\ttoggleSidebarMobile(sidebar, sidebarBackdrop, toggleSidebarMobileHamburger, toggleSidebarMobileClose);\n\t\t\t});\n\t\t}\n    </script></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func headerComponent(title string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "<head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/view/component/workspace/workspace.templ`, Line: 57, Col: 16}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "</title><link rel=\"icon\" type=\"image/svg\" sizes=\"32x32\" href=\"/static/img/pint-of-beer-svgrepo-com.svg\"><link href=\"/static/css/tailwind-output.gen.css\" rel=\"stylesheet\"><link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/choices.js/public/assets/styles/choices.min.css\"><link href=\"/static/css/app.css\" rel=\"stylesheet\"><script src=\"https://unpkg.com/htmx.org@2.0.3\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/htmx.org@1.9.11/dist/ext/path-params.js\"></script><script src=\"https://unpkg.com/htmx-ext-response-targets@2.0.1/response-targets.js\"></script><script src=\"https://unpkg.com/htmx-ext-debug@2.0.0/debug.js\"></script><script src=\"https://cdn.jsdelivr.net/npm/choices.js/public/assets/scripts/choices.min.js\"></script><script src=\"/static/js/select.js\"></script><script src=\"/static/js/upload.js\" type=\"module\"></script><link href=\"https://releases.transloadit.com/uppy/v3.27.0/uppy.min.css\" rel=\"stylesheet\"><script defer src=\"https://cdn.jsdelivr.net/npm/@alpinejs/intersect@3.x.x/dist/cdn.min.js\"></script><script defer src=\"https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js\"></script></head>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
