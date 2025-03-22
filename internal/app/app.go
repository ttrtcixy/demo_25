package application

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	_ "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/ttrtcixy/demo/internal/models"
	"github.com/ttrtcixy/demo/internal/storage"
	"log"
	"strconv"
)

type App struct {
	db  *storage.DB
	app fyne.App
	w   fyne.Window
}

type partnerTable struct {
	partners          *models.Partners
	selectedPartnerID int
	table             *widget.Table
	addButton         *widget.Button
	deleteButton      *widget.Button
}

func NewApp() *App {
	return &App{db: storage.NewDB(),
		app: app.New(),
	}
}

func (a *App) Run() {
	a.w = a.app.NewWindow("Управление Партнерами")

	partnersTable, err := a.PartnersTable()
	if err != nil {
		log.Fatal(err)
	}
	scrollContainer := container.NewHScroll(partnersTable.table)
	scrollContainer.SetMinSize(fyne.NewSize(800, 400))
	tabs := container.NewAppTabs(
		container.NewTabItem("Партнеры", container.NewBorder(nil, container.NewHBox(partnersTable.addButton, partnersTable.deleteButton), nil, nil, scrollContainer)))
	a.w.SetContent(tabs)
	a.w.Resize(fyne.NewSize(1000, 600))
	a.w.ShowAndRun()
}

func (a *App) PartnersTable() (*partnerTable, error) {
	t := partnerTable{}
	var err error

	t.partners, err = a.db.GetPartners()
	if err != nil {
		return nil, err
	}

	table := widget.NewTable(
		func() (int, int) {
			return len(*t.partners) + 1, 6 // +1 для заголовков, 8 колонок (пустой столбец + 7 полей)
		},
		func() fyne.CanvasObject {
			return container.NewHScroll(widget.NewLabel("template"))
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			scrollContainer := o.(*container.Scroll)
			label := scrollContainer.Content.(*widget.Label)
			if i.Row == 0 {
				// Заголовки столбцов
				switch i.Col {
				case 0:
					label.SetText("") // Пустой заголовок для первого столбца
				case 1:
					label.SetText("Название Компании")
				case 2:
					label.SetText("Тип Компании")
				case 3:
					label.SetText("Директор")
				case 4:
					label.SetText("Телефон")
				case 5:
					label.SetText("Рейтинг")
				}
			} else {
				// Данные партнеров
				p := (*t.partners)[i.Row-1]
				switch i.Col {
				case 0:
					label.SetText("")
				case 1:
					label.SetText(p.CompanyName)
				case 2:
					label.SetText(p.PartnerType)
				case 3:
					label.SetText(p.Director)
				case 4:
					label.SetText(p.Phone)
				case 5:
					label.SetText(fmt.Sprintf("%d", p.Rating))
				}
			}
		},
	)

	// Настройка ширины колонок
	table.SetColumnWidth(0, 50)  // Пустой столбец
	table.SetColumnWidth(1, 150) // Name
	table.SetColumnWidth(2, 100) // Type
	table.SetColumnWidth(3, 150) // Director
	table.SetColumnWidth(4, 120) // Phone
	table.SetColumnWidth(5, 200) // Email
	table.SetColumnWidth(6, 200) // Address
	table.SetColumnWidth(7, 80)  // Rating

	t.table = table
	t.selectPartnerColumn(a)
	t.addPartnerButton(a)
	t.deletePartnerButton(a)

	return &t, nil
}

func (t *partnerTable) addPartnerButton(a *App) {
	addButton := widget.NewButton("Add Partner", func() {
		showPartnerForm(a.w, models.Partner{}, func(newPartner models.Partner) {
			err := a.db.AddPartner(newPartner)
			if err != nil {
				dialog.ShowError(err, a.w)
				log.Println(err)
			} else {
				t.partners, err = a.db.GetPartners()
				if err != nil {
					dialog.ShowError(err, a.w)
					log.Println(err)
				}
			}
			t.table.Refresh()
		})
	})
	t.addButton = addButton
}

func (t *partnerTable) deletePartnerButton(a *App) {
	deleteButton := widget.NewButton("Delete Partner", func() {
		if t.selectedPartnerID != 0 {
			err := a.db.DeletePartner(t.selectedPartnerID)
			if err != nil {
				dialog.ShowError(err, a.w)
				log.Println(err)
			}
			t.partners, err = a.db.GetPartners()
			if err != nil {
				dialog.ShowError(err, a.w)
				log.Println(err)
			}
			t.table.Refresh()
			t.selectedPartnerID = 0 // Сброс выбранного партнера
		} else {
			dialog.ShowInformation("No Selection", "Please select a partner to delete", a.w)
		}
	})
	t.deleteButton = deleteButton
}

func (t *partnerTable) selectPartnerColumn(a *App) {
	t.table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 { // Кликабельна вся строка, кроме заголовков
			t.selectedPartnerID = (*t.partners)[id.Row-1].Id
			p := (*t.partners)[id.Row-1]
			showPartnerForm(a.w, p, func(updatedPartner models.Partner) {
				err := a.db.AddPartner(updatedPartner)
				if err != nil {
					dialog.ShowError(err, a.w)
				} else {
					t.partners, err = a.db.GetPartners()
					if err != nil {
						log.Println(err)
					}
					t.table.Refresh()
				}
			})
		}
	}

}

func showPartnerForm(w fyne.Window, p models.Partner, onSave func(models.Partner)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(p.CompanyName)

	typeEntry := widget.NewSelect([]string{"Type 1", "Type 2", "Type 3"}, func(s string) {
		p.PartnerType = s
	})
	typeEntry.SetSelected(p.PartnerType)

	directorEntry := widget.NewEntry()
	directorEntry.SetText(p.Director)

	phoneEntry := widget.NewEntry()
	phoneEntry.SetText(p.Phone)

	emailEntry := widget.NewEntry()
	emailEntry.SetText(p.Email)

	addressEntry := widget.NewEntry()
	addressEntry.SetText(p.Address)

	ratingEntry := widget.NewEntry()
	ratingEntry.SetText(fmt.Sprintf("%d", p.Rating))

	form := widget.NewForm(
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Type", typeEntry),
		widget.NewFormItem("Director", directorEntry),
		widget.NewFormItem("Phone", phoneEntry),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Address", addressEntry),
		widget.NewFormItem("Rating", ratingEntry),
	)

	form.OnSubmit = func() {
		rating, err := strconv.Atoi(ratingEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Rating must be a number"), w)
			return
		}

		p.CompanyName = nameEntry.Text
		p.PartnerType = typeEntry.Selected
		p.Director = directorEntry.Text
		p.Phone = phoneEntry.Text
		p.Email = emailEntry.Text
		p.Address = addressEntry.Text
		p.Rating = rating

		onSave(p)
	}

	dialog.ShowCustomConfirm("Edit Partner", "Save", "Cancel", form, func(b bool) {
		if b {
			form.OnSubmit()
		}
	}, w)
}
