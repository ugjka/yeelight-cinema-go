package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nunows/goyeelight"
)

func main() {
	headFlag := flag.Int("head", 0, "select monitor")
	yeeAddr := flag.String("addr", "", "yeelight adress")
	yeePort := flag.String("port", "55443", "yeelight port")
	flag.Parse()
	if *yeeAddr == "" {
		fmt.Fprintf(os.Stderr, "invalid yeelight adress")
		os.Exit(1)
	}
	yee := goyeelight.New(*yeeAddr, *yeePort)
	log.Println(yee.On())
	c, err := xgb.NewConn()
	if err != nil {
		panic(err)
	}

	err = xinerama.Init(c)
	if err != nil {
		panic(err)
	}

	reply, err := xinerama.QueryScreens(c).Reply()
	if err != nil {
		panic(err)
	}
	if len(reply.ScreenInfo) < *headFlag+1 {
		fmt.Fprintf(os.Stderr, "%d %s\n", *headFlag, "head out of range!")
		os.Exit(1)
	}
	head := reply.ScreenInfo[*headFlag]

	screen := xproto.Setup(c).DefaultScreen(c)
	x0 := head.XOrg
	y0 := head.YOrg
	//Catch CTRL+C
	var interrupt = make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		yee.Off()
		c.Close()
		fmt.Fprintf(os.Stderr, "\nSwitching off\n")
		os.Exit(0)
	}()
	for {
		now := time.Now()
		xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root), x0, y0, head.Width, head.Height, 0xffffffff).Reply()
		if err != nil {
			panic(err)
		}

		data := xImg.Data
		for i := 0; i < len(data); i += 4 {
			data[i], data[i+2], data[i+3] = data[i+2], data[i], 255
		}
		img := &image.RGBA{data, 4 * int(head.Width), image.Rect(0, 0, int(head.Width), int(head.Height))}
		items, err := prominentcolor.Kmeans(img)
		if err != nil {
			panic(err)
		}
		col, _ := colorful.Hex(fmt.Sprintf("#%s", items[0].AsString()))
		h, s, v := col.Hsv()
		changeDur := time.Now().Sub(now).Nanoseconds() / 1000000
		//Minimum change duration is 30ms for yeelight
		if changeDur < 30 {
			yee.SetBright(strconv.FormatFloat(v*100, 'f', 1, 64), "smooth", "30")
			yee.SetHSV(strconv.FormatFloat(h, 'f', 1, 64), strconv.FormatFloat(s*100, 'f', 1, 64), "smooth", "30")
			time.Sleep(30 - time.Duration(changeDur)*time.Millisecond + 5)
		} else {
			yee.SetBright(strconv.FormatFloat(v*100, 'f', 1, 64), "smooth", fmt.Sprintf("%d", changeDur))
			yee.SetHSV(strconv.FormatFloat(h, 'f', 1, 64), strconv.FormatFloat(s*100, 'f', 1, 64), "smooth", fmt.Sprintf("%d", changeDur))
			time.Sleep(time.Millisecond * 5)
		}
		fmt.Printf("\rLatency: %s", time.Now().Sub(now))
	}
}
