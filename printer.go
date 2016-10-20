package main

import (
	"net"

	"github.com/harrydb/go/img/grayscale"
	"github.com/nfnt/resize"

	"fmt"
	"image"
	"log"
	"net/url"
	"os"
	"runtime"

	"pault.ag/go/epson"
	"pault.ag/go/epson/drivers/epspos"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/sourcegraph/go-webkit2/webkit2"
)

func ohshit(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	pageURL := os.Args[2]
	if _, err := url.Parse(pageURL); err != nil {
		log.Fatalf("Failed to parse URL %q: %s", pageURL, err)
	}

	runtime.LockOSThread()
	gtk.Init(nil)
	win, err := gtk.OffscreenWindowNew()
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetDefaultSize(512, 0)

	webView := webkit2.NewWebView()
	defer webView.Destroy()
	win.Add(webView)
	win.ShowAll()

	webView.Connect("load-failed", func() {
		fmt.Println("Load failed.")
	})
	webView.Connect("load-changed", func(_ *glib.Object, loadEvent webkit2.LoadEvent) {
		switch loadEvent {
		case webkit2.LoadFinished:
			fmt.Printf("Loaded, now waiting\n")
			webView.GetSnapshot(func(result *image.RGBA, err error) {
				resizedSrc := resize.Resize(512, 0, result, resize.Lanczos3)
				gray := grayscale.Convert(resizedSrc, grayscale.ToGrayLuminance)

				conn, err := net.Dial("tcp", os.Args[1])
				if err != nil {
					panic(err)
				}

				printer := epspos.New(conn)
				ohshit(printer.Init())
				ohshit(printer.Justification(epson.Center))
				ohshit(printer.PrintImage(*gray))
				ohshit(printer.Feed(4))
				ohshit(printer.Cut())
				gtk.MainQuit()
			})
		}
	})

	glib.IdleAdd(func() bool {
		webView.LoadURI(pageURL)
		return false
	})

	gtk.Main()
}
