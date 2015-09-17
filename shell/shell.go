package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

var output_mode = termbox.OutputNormal

func printString(x, y int, s string, fg termbox.Attribute, bg termbox.Attribute) int {
	for _, r := range s {
		termbox.SetCell(x, y, r, fg, bg)
		x++
	}
	return x
}

func printRightString(x, y int, s string, fg termbox.Attribute, bg termbox.Attribute) int {
	newX := x - len(s)
	return printString(newX, y, s, fg, bg)
}

const horLine = 0x2500

func drawLine(y int, count int) {
	for i := 0; i < count-1; i++ {
		termbox.SetCell(i, y, horLine, termbox.ColorGreen, termbox.ColorDefault)
	}
}

type DrawContext struct {
	width              int
	height             int
	entryLine          int
	folderRightPoint   int
	outputArea         int
	folder             string
	currentInputString string
}

func (d *DrawContext) init() {
	d.currentInputString = ""
	d.folder = "/"
	d.setDimensions(termbox.Size())
}

func (d *DrawContext) setDimensions(width int, height int) {
	d.width = width
	d.height = height
	d.entryLine = height - 2
	d.folderRightPoint = width - 1
	d.outputArea = 3
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	d.printTitle()
	d.printFolder()
	d.drawLines()
	d.drawInputLine()
	termbox.Flush()
}

func (d *DrawContext) printTitle() {
	newCol := printString(0, 0, "Phoenix Meta File System", termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault)
	widthHeight := fmt.Sprintf("(%d,%d)", d.width, d.height)
	newCol = printString(newCol+1, 0, widthHeight, termbox.ColorBlue|termbox.AttrBold, termbox.ColorDefault)
}

func (d *DrawContext) printFolder() {
	folderString := fmt.Sprintf("Folder: %s", d.folder)
	printRightString(d.folderRightPoint, 0, folderString, termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault)
}

func (d *DrawContext) executeLine() {
	d.currentInputString = ""
	d.drawInputLine()
}

func (d *DrawContext) drawLines() {
	drawLine(1, d.width)
	drawLine(d.height-3, d.width)
	drawLine(d.height-1, d.width)
}

func (d *DrawContext) addToInput(c rune) {
	d.currentInputString = fmt.Sprintf("%s%c", d.currentInputString, c)
	d.drawInputLine()
}

func (d *DrawContext) removeInput() {
	if len(d.currentInputString) > 0 {
		d.currentInputString = d.currentInputString[0 : len(d.currentInputString)-1]
		d.drawInputLine()
	}
}

func clearArea(x, y, endx, endy int) {
	for xpoint := x; xpoint < endx; xpoint++ {
		for ypoint := y; ypoint < endy; ypoint++ {
			termbox.SetCell(xpoint, ypoint, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (d *DrawContext) drawInputLine() {
	printString(0, d.entryLine, d.currentInputString, termbox.ColorYellow, termbox.ColorDefault)
	clearArea(len(d.currentInputString), d.entryLine, d.width, d.entryLine+1)
	d.setInputCursor()
}

func (d *DrawContext) setInputCursor() {
	termbox.SetCursor(len(d.currentInputString), d.entryLine)
}

var context DrawContext

func main() {
	err := termbox.Init()

	if err != nil {
		panic(err)
	}
	context.init()
	defer termbox.Close()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break loop
			case termbox.KeyEnter:
				context.executeLine()
			case termbox.KeyBackspace:
				context.removeInput()
			case termbox.KeyBackspace2:
				context.removeInput()
			default:
				context.addToInput(ev.Ch)
			}
			keyString := fmt.Sprintf("Key Event %v", ev)
			printString(0, 10, keyString, termbox.ColorMagenta, termbox.ColorDefault)
			charString := fmt.Sprintf("Key pressed %c", ev.Ch)
			printString(0, 11, charString, termbox.ColorMagenta, termbox.ColorDefault)
			termbox.Flush()
		case termbox.EventResize:
			context.setDimensions(termbox.Size())
		}
	}
}
