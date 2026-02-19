package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	checkmark = `✓`
	crossmark = `✗`
	warnsign  = `⚠`
	upArrow   = `↑`
)

var (
	clrPrimary = lipgloss.Color("208") // orange
	clrAccent  = lipgloss.Color("75")  // sky blue
	clrMuted   = lipgloss.Color("243") // gray
	clrError   = lipgloss.Color("203") // salmon
	clrBorder  = lipgloss.Color("238") // dark gray

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(clrPrimary)
	selectedStyle = lipgloss.NewStyle().Foreground(clrAccent).Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(clrMuted)
	errorStyle    = lipgloss.NewStyle().Foreground(clrError)

	huhTheme = ciergeTheme()
)

func ciergeTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Base = t.Focused.Base.BorderForeground(clrBorder)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(clrPrimary).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(clrPrimary).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(clrMuted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(clrError)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(clrError)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(clrPrimary)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(clrPrimary)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(clrPrimary)
	t.Focused.Option = t.Focused.Option.Foreground(lipgloss.Color("252"))
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(clrPrimary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(clrAccent).Bold(true)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(clrAccent).SetString("[•] ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(lipgloss.Color("252"))
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(clrMuted).SetString("[ ] ")
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("0")).Background(clrPrimary)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(lipgloss.Color("252")).Background(lipgloss.Color("237"))
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(clrAccent)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(clrMuted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(clrPrimary)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}

func runHuh(fields ...huh.Field) error {
	return huh.NewForm(huh.NewGroup(fields...)).WithTheme(huhTheme).Run()
}

func styledTextInput() textinput.Model {
	ti := textinput.New()
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(clrAccent)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(clrMuted)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(clrPrimary)
	return ti
}
