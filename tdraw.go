package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"

	"github.com/mattn/go-runewidth"
)

const (
	MODE_BOX   = "BOX"
	MODE_LINE  = "LINE"
	MODE_TEXT  = "TEXT"
	MODE_ERASE = "ERASE"
)

var defStyle tcell.Style

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func drawLine(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, r rune) {
	// do not draw point
	if x1 == x2 && y1 == y2 {
		return
	}

	cross := false
	if c, _, _, _ := s.GetContent(x1, y1); c != ' ' {
		cross = true
	}

	if x1 == x2 {
		if y1 < y2 {
			for row := y1; row <= y2; row++ {
				s.SetContent(x1, row, '|', nil, tcell.StyleDefault)
			}
			s.SetContent(x1, y2, 'v', nil, tcell.StyleDefault)

		} else {
			for row := y2; row <= y1; row++ {
				s.SetContent(x1, row, '|', nil, tcell.StyleDefault)
			}
			s.SetContent(x1, y2, '^', nil, tcell.StyleDefault)
		}
	} else if y1 == y2 {
		if x1 < x2 {
			for col := x1; col <= x2; col++ {
				s.SetContent(col, y1, '-', nil, tcell.StyleDefault)
			}
			s.SetContent(x2, y1, '>', nil, tcell.StyleDefault)
		} else {
			for col := x2; col <= x1; col++ {
				s.SetContent(col, y1, '-', nil, tcell.StyleDefault)
			}
			s.SetContent(x2, y1, '<', nil, tcell.StyleDefault)
		}
	}

	if cross {
		s.SetContent(x1, y1, '+', nil, tcell.StyleDefault)
	}

	return
}

func drawErase(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, r rune) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		for row := y1; row <= y2; row++ {
			s.SetContent(col, row, ' ', nil, tcell.StyleDefault)
		}
	}
}

func drawBox(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, r rune) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// do not draw point
	if x1 == x2 && y1 == y2 {
		return
	}

	for col := x1 + 1; col < x2; col++ {
		// RuneHLine    = '─'
		s.SetContent(col, y1, tcell.RuneHLine, nil, tcell.StyleDefault)
		s.SetContent(col, y2, tcell.RuneHLine, nil, tcell.StyleDefault)
	}
	for row := y1 + 1; row < y2; row++ {
		// RuneVLine    = '│'
		s.SetContent(x1, row, tcell.RuneVLine, nil, tcell.StyleDefault)
		s.SetContent(x2, row, tcell.RuneVLine, nil, tcell.StyleDefault)
	}

	if y1 == y2 {
		s.SetContent(x1, y1, tcell.RuneHLine, nil, tcell.StyleDefault)
		s.SetContent(x2, y2, tcell.RuneHLine, nil, tcell.StyleDefault)
		return
	}

	if x1 == x2 {
		s.SetContent(x1, y1, tcell.RuneVLine, nil, tcell.StyleDefault)
		s.SetContent(x2, y2, tcell.RuneVLine, nil, tcell.StyleDefault)
		return
	}

	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, tcell.StyleDefault)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, tcell.StyleDefault)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, tcell.StyleDefault)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, tcell.StyleDefault)
	}
}

func drawSelect(s tcell.Screen, x1, y1, x2, y2 int, sel bool) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	for row := y1; row <= y2; row++ {
		for col := x1; col <= x2; col++ {
			mainc, combc, style, width := s.GetContent(col, row)
			if style == tcell.StyleDefault {
				style = defStyle
			}
			style = style.Reverse(sel)
			s.SetContent(col, row, mainc, combc, style)
			col += width - 1
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	encoding.Register()

	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	defer s.Fini()
	defStyle = tcell.StyleDefault
	s.SetStyle(defStyle)

	s.EnableMouse()
	s.Clear()

	// mouse
	mx, my := -1, -1
	ox, oy := -1, -1
	bx, by := -1, -1
	// w, h := s.Size()
	lchar := '*'

	mode := MODE_BOX
	imodeCurX, imodeCurY := 0, 0
	imodeStartX, _ := 0, 0
	// imodeLastC := ' '

loop:
	for {
		// r, _, _, _ := s.GetContent(mx, my)
		emitStr(s, 1, 1, defStyle, fmt.Sprintf("[%3v,%3v] %-7s | esc:box / t:text / l:line / e:erase / MouseR: eraser", mx, my, mode))

		s.Show()
		ev := s.PollEvent()
		st := tcell.StyleDefault
		up := tcell.StyleDefault.
			Background(tcell.ColorBlue).
			Foreground(tcell.ColorBlack)
		// w, h = s.Size()

		// always clear any old selection box
		if ox >= 0 && oy >= 0 && bx >= 0 {
			drawSelect(s, ox, oy, bx, by, false)
		}

		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyCtrlD {
				s.Sync()
				sizeX, sizeY := s.Size()
				arr := make([]string, sizeY)

				for y := 2; y < sizeY; y++ {
					for x := 0; x < sizeX; x++ {
						c, _, _, _ := s.GetContent(x, y)
						arr[y] = arr[y] + string(c)
					}
				}
				{
					n := 0
					for y := 0; y < len(arr); y++ {
						if strings.TrimSpace(arr[y]) == "" {
							n++
						} else {
							break
						}
					}
					arr = arr[n:len(arr)]
				}

				{
					n := 0
					for y := len(arr) - 1; y >= 0; y-- {
						if strings.TrimSpace(arr[y]) == "" {
							n++
						} else {
							break
						}
					}
					arr = arr[0 : len(arr)-n]
				}

				s.Clear()
				for y := 0; y < len(arr); y++ {
					fmt.Println(arr[y])
				}
				break loop
			}

			switch mode {
			case MODE_BOX:
				switch ev.Key() {
				default:
					switch ev.Rune() {
					case 't':
						mode = MODE_TEXT
						imodeCurX, imodeCurY = mx, my
						imodeStartX, _ = mx, my
						s.SetContent(imodeCurX, imodeCurY, '_', nil, st.Blink(true))
					case 'e':
						mode = MODE_ERASE
					case 'l':
						mode = MODE_LINE
					}
				}

			case MODE_ERASE:
				switch ev.Key() {
				case tcell.KeyEscape:
					mode = MODE_BOX
				default:
					switch ev.Rune() {
					case 'e':
						mode = MODE_ERASE
					case 'l':
						mode = MODE_LINE
					case 't', 'i':
						mode = MODE_TEXT
						imodeCurX, imodeCurY = mx, my
						imodeStartX, _ = mx, my
						s.SetContent(imodeCurX, imodeCurY, '_', nil, st.Blink(true))
					}
				}

			case MODE_LINE:
				switch ev.Key() {
				case tcell.KeyEscape:
					mode = MODE_BOX
				default:
					switch ev.Rune() {
					case 'e':
						mode = MODE_ERASE
					case 't', 'i':
						mode = MODE_TEXT
						imodeCurX, imodeCurY = mx, my
						imodeStartX, _ = mx, my
						s.SetContent(imodeCurX, imodeCurY, '_', nil, st.Blink(true))
					}
				}

			case MODE_TEXT:
				switch ev.Key() {
				case tcell.KeyEscape:
					mode = MODE_BOX
					s.SetContent(imodeCurX, imodeCurY, ' ', nil, st)
				case tcell.KeyEnter:
					s.SetContent(imodeCurX, imodeCurY, ' ', nil, st)
					imodeCurX = imodeStartX
					imodeCurY = imodeCurY + 1
				case tcell.KeyDEL:
					if imodeCurX > imodeStartX {
						s.SetContent(imodeCurX, imodeCurY, ' ', nil, st)
						s.SetContent(imodeCurX-1, imodeCurY, '_', nil, st.Blink(true))
						imodeCurX -= 1
					} else {
						s.SetContent(imodeCurX, imodeCurY, '_', nil, st)
					}
				default:
					s.SetContent(imodeCurX, imodeCurY, ev.Rune(), nil, st)
					imodeCurX += runewidth.RuneWidth(ev.Rune())
					s.SetContent(imodeCurX, imodeCurY, '_', nil, st.Blink(true))
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			mx, my = x, y
			button := ev.Buttons()

			if button != tcell.ButtonNone && ox < 0 {
				ox, oy = x, y
			}

			switch mode {

			case MODE_ERASE:
				switch ev.Buttons() {
				case tcell.ButtonNone:
					if ox >= 0 {
						bg := tcell.Color((lchar - '0') * 2)
						drawErase(s, ox, oy, x, y, up.Background(bg), lchar)
						ox, oy = -1, -1
						bx, by = -1, -1
					}
				case tcell.Button1:
					ch := ' '
					bx, by = x, y
					lchar = rune(ch)
					if ox >= 0 && bx >= 0 {
						drawSelect(s, ox, oy, bx, by, true)
					}
				case tcell.Button3:
					ox, oy = -1, -1
					bx, by = -1, -1
					s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
				}

			case MODE_BOX:
				switch ev.Buttons() {
				case tcell.ButtonNone:
					if ox >= 0 {
						bg := tcell.Color((lchar - '0') * 2)
						drawBox(s, ox, oy, x, y, up.Background(bg), lchar)
						ox, oy = -1, -1
						bx, by = -1, -1
					}
				case tcell.Button1:
					ch := ' '
					bx, by = x, y
					lchar = rune(ch)
					if ox >= 0 && bx >= 0 {
						drawSelect(s, ox, oy, bx, by, true)
					}
				case tcell.Button3:
					ox, oy = -1, -1
					bx, by = -1, -1
					s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
				}
			case MODE_LINE:
				switch ev.Buttons() {
				case tcell.ButtonNone:
					if ox >= 0 {
						bg := tcell.Color((lchar - '0') * 2)
						drawLine(s, ox, oy, x, y, up.Background(bg), lchar)
						ox, oy = -1, -1
						bx, by = -1, -1
					}
				case tcell.Button1:
					ch := ' '
					bx, by = x, y
					lchar = rune(ch)
					if ox >= 0 && bx >= 0 {
						drawSelect(s, ox, oy, bx, by, true)
					}
				case tcell.Button3:
					ox, oy = -1, -1
					bx, by = -1, -1
					s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
				}
			}

		}
	}
}
