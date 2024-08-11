package handler

import (
	"fmt"

	"github.com/my-pet-projects/collection/internal/view/page"
	"github.com/my-pet-projects/collection/internal/web"
)

func (h WorkspaceHandler) HandleBeerImagesIndex(reqResp *web.ReqRespPair) error {
	fmt.Println("asdasda")
	// p := page.BeerImagesPage{
	// 	Title: "Beer Images",
	// }
	return reqResp.Render(page.BeerImagesPage())
}
