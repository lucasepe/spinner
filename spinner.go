package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/lucasepe/spinner/pkg/env"
)

const (
	// 150ms per frame
	DEFAULT_FRAME_RATE = time.Millisecond * 150
)

//var DefaultCharset = []string{"|", "/", "-", "\\"}
var DefaultCharset = []string{
	"⠈⠁", "⠈⠑", "⠈⠱", "⠈⡱", "⢀⡱", "⢄⡱", "⢄⡱", "⢆⡱",
	"⢎⡱", "⢎⡰", "⢎⡠", "⢎⡀", "⢎⠁", "⠎⠁", "⠊⠁",
}

type Spinner struct {
	sync.Mutex
	Text      string
	Charset   []string
	FrameRate time.Duration
	runChan   chan struct{}
	stopOnce  sync.Once
	Output    io.Writer
	NoTty     bool
}

// create spinner object
func NewSpinner(title string) *Spinner {
	sp := &Spinner{
		Text:      title,
		Charset:   DefaultCharset,
		FrameRate: DEFAULT_FRAME_RATE,
		runChan:   make(chan struct{}),
	}
	if !env.IsTerminal(os.Stderr) {
		sp.NoTty = true
	}
	return sp
}

// start a new spinner, title can be an empty string
func StartNew(title string) *Spinner {
	return NewSpinner(title).Start()
}

// start spinner
func (sp *Spinner) Start() *Spinner {
	go sp.writer()
	return sp
}

// set custom spinner frame rate
func (sp *Spinner) SetText(s string) *Spinner {
	sp.Lock()
	sp.Text = s
	sp.Unlock()
	return sp
}

// set custom spinner frame rate
func (sp *Spinner) SetSpeed(rate time.Duration) *Spinner {
	sp.Lock()
	sp.FrameRate = rate
	sp.Unlock()
	return sp
}

// set custom spinner character set
func (sp *Spinner) SetCharset(chars []string) *Spinner {
	sp.Lock()
	sp.Charset = chars
	sp.Unlock()
	return sp
}

// stop and clear the spinner
func (sp *Spinner) Stop() {
	//prevent multiple calls
	sp.stopOnce.Do(func() {
		close(sp.runChan)
		sp.clearLine()
		// show cursor
		fmt.Print("\033[?25h")
	})
}

// spinner animation
func (sp *Spinner) animate() {
	var out string
	for i := 0; i < len(sp.Charset); i++ {
		select {
		case <-sp.runChan:
			return
		default:
			out = sp.Charset[i] + " " + sp.Text
			switch {
			case sp.Output != nil:
				fmt.Fprint(sp.Output, out)
			case !sp.NoTty:
				fmt.Print(out)
			}
			time.Sleep(sp.FrameRate)
			sp.clearLine()
		}
	}
}

// write out spinner animation until runChan is closed
func (sp *Spinner) writer() {
	// hide cursor
	fmt.Print("\033[?25l")
	sp.animate()
	for {
		select {
		case <-sp.runChan:
			return
		default:
			sp.animate()
		}
	}
}

// workaround for Mac OS < 10 compatibility
func (sp *Spinner) clearLine() {
	if !sp.NoTty {
		fmt.Printf("\033[2K")
		fmt.Println()
		fmt.Printf("\033[1A")
	}
}
