package internal

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mkozjak/tview"
)

type ModalInput struct {
	*tview.Form
	frame *tview.Frame
	done  func(string, string, bool)
}

func CreateModalInput(listen func(event *tcell.EventKey) *tcell.EventKey) *ModalInput {
	form := tview.NewForm()

	form.SetInputCapture(listen).
		SetBorder(true).
		SetTitle("[::b]Search Tidal").
		SetBorderColor(tcell.ColorCornflowerBlue).
		SetBackgroundColor(tcell.ColorDefault).
		SetCustomBorders(CustomBorders)

	form.SetCancelFunc(func() {
	})

	m := &ModalInput{form, tview.NewFrame(form), nil}

	return m
}

func (m *ModalInput) SetCancelFunc(cb func()) {
	m.Form.SetCancelFunc(cb)
}
