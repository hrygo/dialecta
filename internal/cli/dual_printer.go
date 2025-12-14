package cli

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/term"
)

// DualPrinter handles side-by-side streaming output
type DualPrinter struct {
	mu    sync.Mutex
	out   *os.File
	width int

	proLines []string
	conLines []string

	// Current print state
	printedLines int

	// Configuration
	colWidth int
	gap      int

	leftColor  string
	rightColor string
}

func NewDualPrinter(out *os.File) *DualPrinter {
	w, _, err := term.GetSize(int(out.Fd()))
	if err != nil {
		w = 80 // Default fallback
	}

	// Safety for very narrow screens
	if w < 40 {
		w = 40
	}

	gap := 4
	colW := (w - gap) / 2

	return &DualPrinter{
		out:        out,
		width:      w,
		gap:        gap,
		colWidth:   colW,
		proLines:   []string{""}, // Start with empty line
		conLines:   []string{""}, // Start with empty line
		leftColor:  ColorBrightGreen,
		rightColor: ColorBrightRed,
	}
}

func (dp *DualPrinter) Start() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Print Headers
	headers := dp.formatDualLine("🟢 正方 (PRO)", "🔴 反方 (CON)", ColorBrightGreen, ColorBrightRed)
	fmt.Fprintln(dp.out, headers)
	fmt.Fprintln(dp.out, dp.formatDualLine(strings.Repeat("─", dp.colWidth), strings.Repeat("─", dp.colWidth), ColorDim, ColorDim))

	dp.printedLines = 0
}

func (dp *DualPrinter) UpdatePro(text string) {
	dp.update(true, text)
}

func (dp *DualPrinter) UpdateCon(text string) {
	dp.update(false, text)
}

func (dp *DualPrinter) update(isPro bool, text string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	runes := []rune(text)
	for _, r := range runes {
		dp.appendRune(isPro, r)
	}
	dp.render()
}

func (dp *DualPrinter) appendRune(isPro bool, r rune) {
	var currentLineIdx int
	var currentLineContent string

	if isPro {
		currentLineIdx = len(dp.proLines) - 1
		currentLineContent = dp.proLines[currentLineIdx]
	} else {
		currentLineIdx = len(dp.conLines) - 1
		currentLineContent = dp.conLines[currentLineIdx]
	}

	if r == '\n' {
		if isPro {
			dp.proLines = append(dp.proLines, "")
		} else {
			dp.conLines = append(dp.conLines, "")
		}
		return
	}

	rw := runeWidth(r)
	curW := stringWidth(currentLineContent)

	if curW+rw > dp.colWidth {
		if isPro {
			dp.proLines = append(dp.proLines, string(r))
		} else {
			dp.conLines = append(dp.conLines, string(r))
		}
	} else {
		if isPro {
			dp.proLines[currentLineIdx] += string(r)
		} else {
			dp.conLines[currentLineIdx] += string(r)
		}
	}
}

func (dp *DualPrinter) render() {
	targetHeight := len(dp.proLines)
	if len(dp.conLines) > targetHeight {
		targetHeight = len(dp.conLines)
	}

	currentMaxHeight := targetHeight

	// Scroll if needed
	needed := currentMaxHeight - dp.printedLines
	if needed > 0 {
		fmt.Fprint(dp.out, strings.Repeat("\n", needed))
		dp.printedLines = currentMaxHeight
	}

	// Update visible lines
	// To minimize cursor jumping, we only update the active lines (the last ones)
	// But simply always updating the last line of both is safe enough for append-only logic.
	dp.updateLineOnScreen(len(dp.proLines)-1, true)
	dp.updateLineOnScreen(len(dp.conLines)-1, false)
}

func (dp *DualPrinter) updateLineOnScreen(lineIdx int, isPro bool) {
	if lineIdx < 0 {
		return
	}

	up := dp.printedLines - 1 - lineIdx
	if up < 0 {
		return
	}

	// Move Up
	if up > 0 {
		fmt.Fprintf(dp.out, "\033[%dA", up)
	}

	text := ""
	if isPro {
		text = dp.proLines[lineIdx]
	} else {
		text = dp.conLines[lineIdx]
	}

	// Return to start of line
	fmt.Fprint(dp.out, "\r")

	if isPro {
		fmt.Fprintf(dp.out, "%s%s%s", dp.leftColor, text, ColorReset)
	} else {
		moveRight := dp.colWidth + dp.gap
		fmt.Fprintf(dp.out, "\033[%dC", moveRight)
		fmt.Fprintf(dp.out, "%s%s%s", dp.rightColor, text, ColorReset)
	}

	// Restore Down
	if up > 0 {
		fmt.Fprintf(dp.out, "\033[%dB", up)
	}

	// Cleanup return
	fmt.Fprint(dp.out, "\r")
}

func (dp *DualPrinter) formatDualLine(left, right, cLeft, cRight string) string {
	padLeft := dp.colWidth - stringWidth(left)
	if padLeft < 0 {
		padLeft = 0
	}

	return fmt.Sprintf("%s%s%s%s%s%s%s%s",
		cLeft, left, strings.Repeat(" ", padLeft), ColorReset,
		strings.Repeat(" ", dp.gap),
		cRight, right, ColorReset,
	)
}

func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	if utf8.RuneLen(r) > 1 {
		return 2
	}
	return 1
}
