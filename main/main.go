package main

import (
	"chrome_render/chrome"
	"chrome_render/config"
	"chrome_render/wsserver"
	"context"
	"fmt"
)

var actionFunCtx  context.Context

func main()  {
	ch  := make(chan error, 1)
	go neChrome()
	go wsserver.Start()
	<-ch
}

func neChrome() {
	conf :=config.GetConfig()
	fmt.Print(conf)
	parent := context.Background()
	//conf := config.GetConfig()
	chPageFrames := make(chan *chrome.PageScreencastFrameImage, *conf.ForceFrameRate)

	chrome.NewInstance(parent, "abc", "http://www.kktv5.com/", 1280,720,chPageFrames)
	//defaultOptions := []chromedp.ExecAllocatorOption{
	//	chromedp.NoFirstRun,
	//	chromedp.NoDefaultBrowserCheck,
	//
	//	// After Puppeteer's default behavior.
	//	chromedp.Flag("disable-background-networking", true),
	//	chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
	//	chromedp.Flag("disable-background-timer-throttling", true),
	//	chromedp.Flag("disable-backgrounding-occluded-windows", true),
	//	chromedp.Flag("disable-breakpad", true),
	//	chromedp.Flag("disable-client-side-phishing-detection", true),
	//	chromedp.Flag("disable-default-apps", true),
	//	chromedp.Flag("disable-dev-shm-usage", true),
	//	//chromedp.Flag("disable-extensions", true),
	//	chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
	//	chromedp.Flag("disable-hang-monitor", true),
	//	chromedp.Flag("disable-ipc-flooding-protection", true),
	//	chromedp.Flag("disable-popup-blocking", true),
	//	chromedp.Flag("disable-prompt-on-repost", true),
	//	chromedp.Flag("disable-renderer-backgrounding", true),
	//	chromedp.Flag("disable-sync", true),
	//	chromedp.Flag("force-color-profile", "srgb"),
	//	chromedp.Flag("metrics-recording-only", true),
	//	chromedp.Flag("safebrowsing-disable-auto-update", true),
	//	chromedp.Flag("enable-automation", true),
	//	chromedp.Flag("password-store", "basic"),
	//	chromedp.Flag("use-mock-keychain", true),
	//	chromedp.Flag("hide-scrollbars", true),
	//
	//	chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
	//	chromedp.Flag("load-extension", "/Users/chenfei/Desktop/chrome_render/crx"),
	//	chromedp.Flag("whitelisted-extension-id", "efjphpadcohhfnlcfjbdiehlnhkomdck"),
	//	chromedp.Flag("window-size", fmt.Sprintf("%d,%d", 1280, 720)),
	//
	//	chromedp.Flag("disable-web-security", true),
	//}
	//
	//ctx, _ := chromedp.NewExecAllocator(parent, defaultOptions...)
	//
	//ctx, _ = chromedp.NewContext(ctx,
	//	chromedp.WithDebugf(devToolHandler),
	//	chromedp.WithErrorf(devToolHandler),
	//	chromedp.WithLogf(devToolHandler))
	//
	//err := chromedp.Run(ctx, makeTasks())
	//if err != nil {
	//	panic(err)
	//}
}

//func makeTasks() chromedp.Tasks {
//	//var res int
//	return chromedp.Tasks{
//		chromedp.Navigate("https://www.kktv5.com"),
//		//chromedp.Evaluate(injectJSCodes(i.url), &res),
//		chromedp.ActionFunc(func(ctx context.Context) error {
//			actionFunCtx = ctx
//			return nil
//		}),
//	}
//}
//
//func devToolHandler(s string, is ...interface{}) {
//	fmt.Printf(s, is)
//}