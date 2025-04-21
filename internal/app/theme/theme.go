package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type CustomTheme struct{}

func NewTheme() fyne.Theme {
	return &CustomTheme{}
}

func (m *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.White
	case theme.ColorNameButton, theme.ColorNameInputBackground:
		return color.RGBA{R: 0xF4, G: 0xE8, B: 0xD3, A: 0xFF}
	case theme.ColorNamePrimary, theme.ColorNameFocus:
		return color.RGBA{R: 0x67, G: 0xBA, B: 0x80, A: 0xFF}
	case theme.ColorNameForeground:
		return color.Black
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
