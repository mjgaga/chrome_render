package chrome

import (
	"chrome_render/config"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PageScreencastFrameImage []byte

//type chromeStatus int
//const (
//	chromeStatusError chromeStatus = iota
//	chromeStatusStarting
//	chromeStatusRunning
//	chromeStatusStopped
//)c

var conf = config.GetConfig()

type chromeInstance struct {
	taskName   string
	url        string
	widthSize  int
	heightSize int

	lastFrameData  *PageScreencastFrameImage
	lastFrameTime  time.Time
	chanPageFrames chan<- *PageScreencastFrameImage

	startCtx     context.Context
	actionFunCtx context.Context

	isDoneFlag bool
	pid        int
}

type ChromeRunCallBack struct {
	PID int
	Err error
}

func (i *chromeInstance) Description() string {
	return fmt.Sprintf("taskName = %s, url = %s, wideth = %d, height = %d", i.taskName, i.url, i.widthSize, i.heightSize)
}

func (i *chromeInstance) isDone() bool {
	return i.isDoneFlag
}

func (i *chromeInstance) Start(parent context.Context) error {
	if err:=os.Setenv("DISPLAY", ":2"); err != nil {
		panic(err)
	}

	ctx, _ := chromedp.NewExecAllocator(parent, i.DefaultOptions()...)
	//go func() {
	//	select {
	//	case <-parent.done():
	//		cancel()
	//	}
	//}()

	ctx, _ = chromedp.NewContext(ctx,
		chromedp.WithDebugf(i.devToolHandler),
		chromedp.WithErrorf(i.devToolHandler),
		chromedp.WithLogf(i.devToolHandler))
	//defer cancel()

	err := chromedp.Run(ctx, i.makeTasks())
	if err != nil {
		logrus.WithError(err).Errorf("chromedp run tasks error. task name: %s", i.taskName)
		return err
	}

	return nil
}

func (i *chromeInstance) done() {
	i.isDoneFlag = true
}

func (i *chromeInstance) onPageLoadFired() {
	go func() {
		if i.actionFunCtx == nil {
			logrus.Warnf("%s actionFunCtx is Null, re-try after 100 Millisecond. task name: %s", i.url, i.taskName)
			time.Sleep(time.Millisecond * 100)
			i.onPageLoadFired()
			return
		}

		var res = -1
		ctx, _ := context.WithCancel(i.actionFunCtx)

		if err := chromedp.Evaluate(injectJSCodes(i.url), &res).Do(ctx); err != nil {
			logrus.WithError(err).Errorf("injectAudio on reloaded failed. task name: %s", i.taskName)
			return
		}

		logrus.Printf("%s page onPageLoadFired. task name: %s", i.url, i.taskName)
	}()
}

func (i *chromeInstance) makeTasks() chromedp.Tasks {
	//var res int
	return chromedp.Tasks{
		emulation.SetDeviceMetricsOverride(int64(i.widthSize), int64(i.heightSize), 1.0, false),
		chromedp.Navigate(i.url),
		//chromedp.Evaluate(injectJSCodes(i.url), &res),
		chromedp.ActionFunc(func(ctx context.Context) error {
			i.actionFunCtx = ctx
			return nil
		}),
	}
}

func (i *chromeInstance) onPageScreencastFrame(params []byte) {
	if i.isDoneFlag {
		logrus.Errorf("chrome (url: %s) instance stopped. task name: %s", i.url, i.taskName)
		return
	}

	var psf page.EventScreencastFrame

	if err := psf.UnmarshalJSON(params); err != nil {
		logrus.WithError(err).Errorf("cant not unmarshal JSON. task name: %s", i.taskName)
		return
	}

	//if err := page.ScreencastFrameAck(psf.SessionID).Do(i.actionFunCtx); err != nil {
	//	logrus.WithError(err).Warnln("cant not unmarshal JSON")
	//}

	jpgData, err := base64.StdEncoding.DecodeString(psf.Data)
	if err != nil {
		logrus.WithError(err).Errorln("cant not decode base64 string to jpeg data. task name: %s", i.taskName)
		jpgData = []byte{}
	}

	a := PageScreencastFrameImage(jpgData)
	i.lastFrameData = &a
	i.lastFrameTime = time.Now()

	go func() {
		if *conf.SaveFrameJpg {
			dirPath := fmt.Sprintf("%s/%s", *conf.FrameJpgPath, i.taskName)
			if !exists(dirPath) {
				os.MkdirAll(dirPath, 0755)
			}

			fileName := fmt.Sprintf("%s/%d.jpg", dirPath, time.Now().UnixNano())
			err = ioutil.WriteFile(fileName, jpgData, 0644)
			if err != nil {
				logrus.Errorln(err)
			}
		}
	}()
}

func (i *chromeInstance) devToolHandler(s string, is ...interface{}) {
	logrus.Tracef(s, is)

	for _, elem := range is {
		var msg cdproto.Message
		err := json.Unmarshal([]byte(fmt.Sprintf("%s", elem)), &msg)
		if err != nil {
			continue
		}

		switch msg.Method {
		default:
			logrus.Debugf("method: %s", msg.Method)
		case cdproto.EventPageLoadEventFired:
			i.onPageLoadFired()
		case cdproto.EventPageScreencastFrame:
			i.onPageScreencastFrame(msg.Params)
		case cdproto.EventRuntimeConsoleAPICalled:
			i.onPageConsole(msg.Params)
		}
	}
}

func (i *chromeInstance) onPageConsole(params []byte) {
	consoleEvent := runtime.EventConsoleAPICalled{}
	if err := consoleEvent.UnmarshalJSON(params); err != nil {
		logrus.WithError(err).Printf("unmarshal ConsoleAPICalled event failed. task name: %s", i.taskName)
		return
	}

	for _, a := range consoleEvent.Args {
		logrus.Debugf("task name: %s [%s] [%s] %s", i.taskName, consoleEvent.Timestamp.Time(), consoleEvent.Type, a.Description)
	}
}

func (i *chromeInstance) DefaultOptions() []chromedp.ExecAllocatorOption {
	defaultOptions := append(chromedp.DefaultExecAllocatorOptions[:], i.headless)
	return defaultOptions
}

func (i *chromeInstance) headless(a *chromedp.ExecAllocator) {
	chromedp.Flag("headless", false)(a)
	chromedp.Flag("mute-audio", false)(a)
	chromedp.Flag("hide-scrollbars", true)(a)
	chromedp.Flag("start-maximized", true)(a)
	//chromedp.Flag("kiosk", true)(a)

	chromedp.Flag("autoplay-policy", "no-user-gesture-required")(a)
	chromedp.Flag("auto-open-devtools-for-tabs", true)(a)

	chromedp.Flag("disable-extensions", false)(a)
	chromedp.Flag("load-extension", "/Users/chenfei/Desktop/chrome_render/crx")(a)
	chromedp.Flag("whitelisted-extension-id", "efjphpadcohhfnlcfjbdiehlnhkomdck")(a)
	//chromedp.Flag("whitelisted-extension-id", "dbhgfmbemebmbccbpibhhoeddnpicikl")(a)

	chromedp.Flag("window-size", fmt.Sprintf("%d,%d", i.widthSize, i.heightSize))(a)
	chromedp.Flag("disable-web-security", true)(a)
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func injectJSCodes(url string) string {
	conf := config.GetConfig()
	defer logrus.Printf("[%s] audio injected.", url)

	var base64Str string
	if *conf.InjectNoise {
		base64Str = `"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";`
	} else {
		base64Str = `"AAAAHGZ0eXBNNEEgAAAAAE00QSBtcDQyaXNvbQAAAAFtZGF0AAAAAAAAAC4hAANAaBwhAANAaBwhAANAaBwhAANAaBwhAANAaBwAAANGbW9vdgAAAGxtdmhkAAAAANqJYlDaiWJQAACsRAAACGQAAQAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAdh0cmFrAAAAXHRraGQAAAAB2oliUNqJYlAAAAABAAAAAAAACGQAAAAAAAAAAAAAAAABAAAAAAEAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAF0bWRpYQAAACBtZGhkAAAAANqJYlDaiWJQAACsRAAAFABVxAAAAAAAMWhkbHIAAAAAAAAAAHNvdW4AAAAAAAAAAAAAAABDb3JlIE1lZGlhIEF1ZGlvAAAAARttaW5mAAAAEHNtaGQAAAAAAAAAAAAAACRkaW5mAAAAHGRyZWYAAAAAAAAAAQAAAAx1cmwgAAAAAQAAAN9zdGJsAAAAZ3N0c2QAAAAAAAAAAQAAAFdtcDRhAAAAAAAAAAEAAAAAAAAAAAACABAAAAAArEQAAAAAADNlc2RzAAAAAAOAgIAiAAAABICAgBRAFQAYAAAB9AAAAfQABYCAgAISEAaAgIABAgAAABhzdHRzAAAAAAAAAAEAAAAFAAAEAAAAABxzdHNjAAAAAAAAAAEAAAABAAAABQAAAAEAAAAoc3RzegAAAAAAAAAAAAAABQAAAAYAAAAGAAAABgAAAAYAAAAGAAAAFHN0Y28AAAAAAAAAAQAAACwAAAD6dWR0YQAAAPJtZXRhAAAAAAAAACJoZGxyAAAAAAAAAABtZGlyAAAAAAAAAAAAAAAAAAAAAADEaWxzdAAAALwtLS0tAAAAHG1lYW4AAAAAY29tLmFwcGxlLmlUdW5lcwAAABRuYW1lAAAAAGlUdW5TTVBCAAAAhGRhdGEAAAABAAAAACAwMDAwMDAwMCAwMDAwMDg0MCAwMDAwMDM1QiAwMDAwMDAwMDAwMDAwODY1IDAwMDAwMDAwIDAwMDAwMDAwIDAwMDAwMDAwIDAwMDAwMDAwIDAwMDAwMDAwIDAwMDAwMDAwIDAwMDAwMDAwIDAwMDAwMDAw";`
	}

	var jsCodes = []string{
		`(()=>{`,
		`function b64toBlob(b64Data, contentType='', sliceSize=512) {
  			const byteCharacters = atob(b64Data), byteArrays = [];
  			for (let offset = 0; offset < byteCharacters.length; offset += sliceSize) {
    		const slice = byteCharacters.slice(offset, offset + sliceSize);
    		const byteNumbers = new Array(slice.length);
    		for (let i = 0; i < slice.length; i++) {
      			byteNumbers[i] = slice.charCodeAt(i);
    		}
    		const byteArray = new Uint8Array(byteNumbers);
    		byteArrays.push(byteArray);
		}
  		return new Blob(byteArrays, {type: contentType});}`,

		`const label = document.createElement("span");
        label.style.fontSize = "14px";
        label.style.position = "fixed";
		label.style.zIndex = "10000";
        label.style.bottom = "5px";
        label.style.right = "5px";
        label.style.color = "rgba(200, 200, 200, 0.8)";
        setInterval(() => {
            var n = new Date();
            label.innerText = n.getFullYear()+"/"+(n.getMonth()+1)+"/"+n.getDate()+" "+n.getHours()+":"+n.getMinutes()+":"+n.getSeconds();
        }, 1000);
        document.body.append(label);`,

		fmt.Sprintf(`const base64 = %s`, base64Str),
		`const blob = b64toBlob(base64, "audio/x-m4a");
		const blobUrl = URL.createObjectURL(blob);
		const muteAudio = new Audio;
		muteAudio.loop = true;
		muteAudio.src = blobUrl
		muteAudio.play();
		return 0;`,
		`})();`,
	}

	jsCodeStr := strings.Join(jsCodes, ";")
	logrus.Debugf("inject js code: %s", jsCodeStr)
	return jsCodeStr
}

func NewInstance(ctx context.Context, taskName, url string, widthSize, heightSize int, chPageFrames chan<- *PageScreencastFrameImage) {
	//ctx, cancel := context.WithCancel(parentCtx)

	var chrome = chromeInstance{
		taskName:       taskName,
		url:            url,
		widthSize:      widthSize,
		heightSize:     heightSize,
		chanPageFrames: chPageFrames,
		isDoneFlag:     false,
	}

	logrus.Printf("new chrome browser: %s", chrome.Description())

	chrome.Start(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				logrus.Printf("%s task will done", chrome.taskName)
				chrome.done()
				return
			}
		}
	}()
}

func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
