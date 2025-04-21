package application

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/ttrtcixy/demo/internal/app/theme"
	"github.com/ttrtcixy/demo/internal/storage"
	"log"
)

type App struct {
	db    *storage.DB
	app   fyne.App
	w     fyne.Window
	theme fyne.Theme
}

func NewApp() *App {
	return &App{
		db:    storage.NewDB(),
		app:   app.New(),
		theme: theme.NewTheme(),
	}
}

func (a *App) LoadTheme() {
	iconResource, err := fyne.LoadResourceFromPath("./icon.ico")
	if err != nil {
		log.Println(err)
	}
	a.w.SetIcon(iconResource)

	a.app.Settings().SetTheme(a.theme)
}

func (a *App) InitTabs() *container.AppTabs {
	partnersTable, err := a.partnersTable()
	if err != nil {
		log.Println(err)
	}

	if partnersTable.partners == nil {
		dialog.ShowInformation("Нет данных", "Партнеры не найдены. Добавьте нового партнера.", a.w)
	}

	scrollContainer := container.NewHScroll(partnersTable.table)
	scrollContainer.SetMinSize(fyne.NewSize(800, 400))

	tabs := container.NewAppTabs(
		container.NewTabItem("Партнеры", container.NewBorder(
			nil,
			container.NewHBox(partnersTable.addButton, partnersTable.deleteButton),
			nil, nil,
			scrollContainer,
		)),
		container.NewTabItem("Продажи", a.createSalesTab()),
		container.NewTabItem("Расчет материалов", a.createMaterialsCalcTab()),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

func (a *App) Run() {

	a.w = a.app.NewWindow("Управление Партнерами")

	a.LoadTheme()

	tabs := a.InitTabs()

	a.w.SetContent(tabs)
	a.w.Resize(fyne.NewSize(1200, 600))
	a.w.ShowAndRun()
}
