package bendis

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"log"
	"time"
)

func (b *Bendis) TakeScreenshot(pageURL, testName string, width, height float64) {
	page := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage(pageURL).MustWaitLoad()

	img, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
		Clip: &proto.PageViewport{
			X:      0,
			Y:      0,
			Width:  width,
			Height: height,
			Scale:  1,
		},
		FromSurface: true,
	})
	if err != nil {
		log.Println(err)
	}
	fileName := time.Now().Format("02-01-2006-15-04-05.000000")
	_ = utils.OutputFile(fmt.Sprintf("%s/screenshots/%s-%s.png", b.RootPath, testName, fileName), img)
}

func (b *Bendis) FetchPage(pageURL string) *rod.Page {
	return rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage(pageURL).MustWaitLoad()
}

func (b *Bendis) SelectElementByID(page *rod.Page, id string) *rod.Element {
	return page.MustElementByJS(fmt.Sprintf("document.getElementById('%s')", id))
	// add other elements here
}
