package cpu

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sshaman1101/uvm/defines"
)

type videoCard struct {
	mem [defines.VideoWidth * defines.VideoHeight]uint8

	display *image.NRGBA
	redrawn chan struct{}
}

// todo:
// it just a POC, more to fix and reconsider here:
// * docstrings
// * do not touch host's file system
// * we need mem-mapped IO, so videoCard.mem have to be refactored anyway

func newVideo() *videoCard {
	if err := os.RemoveAll("/tmp/video/"); err != nil {
		panic(err)
	}
	if err := os.Mkdir("/tmp/video/", 0o755); err != nil {
		panic(err)
	}

	buf := image.NewNRGBA(image.Rect(0, 0, defines.VideoWidth, defines.VideoHeight))

	black := color.NRGBA{
		A: 255,
		R: 0,
		G: 0,
		B: 0,
	}

	for y := 0; y < defines.VideoHeight; y++ {
		for x := 0; x < defines.VideoWidth; x++ {
			buf.Set(x, y, black)
		}
	}

	video := &videoCard{
		display: buf,
		mem:     [defines.VideoWidth * defines.VideoHeight]uint8{},
		redrawn: make(chan struct{}),
	}

	go video.startWebSocket()
	return video
}

func (v *videoCard) render() {
	for x := 0; x < defines.VideoWidth; x++ {
		for y := 0; y < defines.VideoHeight; y++ {
			color8 := v.mem[x*defines.VideoWidth+y]
			v.point(x, y, color8)
		}
	}
}

const lastFrameAt = "/tmp/video/last.png"

var map3 = [8]uint8{
	0, 37, 74, 110, 146, 183, 219, 255,
}

var map2 = [4]uint8{
	0, 85, 170, 255,
}

func (v *videoCard) point(x, y int, color8 uint8) {
	r := (color8 >> 5) & 0x07
	r = map3[r]

	g := (color8 >> 2) & 0x07
	g = map3[g]

	b := color8 & 0x03
	b = map2[b]

	v.display.Set(x, y, color.NRGBA{
		A: 255, R: r, G: g, B: b,
	})
}

func (v *videoCard) writeFrame() {
	// it very unlikely will cause any performance issues,
	// but consider encode to a bytes buffer once,
	// then write such buffer to each file.
	fd, err := os.Create(lastFrameAt)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(fd, v.display); err != nil {
		fd.Close()
		log.Fatal(err)
	}

	if err := fd.Close(); err != nil {
		log.Fatal(err)
	}

	select {
	case v.redrawn <- struct{}{}:
	default:
	}
}

// below is a websocket-as-a-display implementation

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	filePeriod = 20 * time.Millisecond
	writeWait  = filePeriod
)

func (v *videoCard) reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (v *videoCard) writer(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(2 * time.Second) // alpha-like frame?
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()
	readAndSend := func() {
		bs, err := os.ReadFile(lastFrameAt)
		if err != nil {
			log.Printf("failed to read last frame: %v", err)
			return
		}

		ws.SetWriteDeadline(time.Now().Add(writeWait))
		if err := ws.WriteMessage(websocket.BinaryMessage, bs); err != nil {
			log.Printf("failed to write: %v", err)
			return
		}
	}
	for {
		select {
		case <-fileTicker.C:
			readAndSend()
		case <-v.redrawn:
			readAndSend()
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (v *videoCard) serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go v.writer(ws)
	v.reader(ws)
}

func (v *videoCard) serveIndex(w http.ResponseWriter, r *http.Request) {
	// TODO(nikonov): use embed
	// TODO(nikonov): implement websocket reconnet on a client
	bs, err := os.ReadFile("assets/index.html")
	if err != nil {
		panic(err)
	}

	w.Write(bs)
}

func (v *videoCard) startWebSocket() {
	http.HandleFunc("/", v.serveIndex)
	http.HandleFunc("/ws", v.serveWs)

	log.Printf("starting websocket server at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
