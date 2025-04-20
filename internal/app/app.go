package application

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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
	return &App{
		db:  storage.NewDB(),
		app: app.New(),
	}
}

func (a *App) Run() {
	iconResource, err := fyne.LoadResourceFromPath("./cmd/icon.png")

	a.w = a.app.NewWindow("Управление Партнерами")
	a.w.SetIcon(iconResource)

	partnersTable, err := a.PartnersTable()
	if err != nil {
		log.Println(err)
	}

	// Проверка на отсутствие партнеров
	if partnersTable.partners == nil {
		dialog.ShowInformation("Нет данных", "Партнеры не найдены. Добавьте нового партнера.", a.w)
	}

	scrollContainer := container.NewHScroll(partnersTable.table)
	scrollContainer.SetMinSize(fyne.NewSize(800, 400))
	//tabs := container.NewAppTabs(
	//	container.NewTabItem("Партнеры", container.NewBorder(nil, container.NewHBox(partnersTable.addButton, partnersTable.deleteButton), nil, nil, scrollContainer)),
	//)
	// Создаем содержимое для новой вкладки (пока заглушка)

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

	a.w.SetContent(tabs)
	a.w.Resize(fyne.NewSize(1200, 600))
	a.w.ShowAndRun()
}

func (a *App) PartnersTable() (*partnerTable, error) {
	t := &partnerTable{}
	var err error

	t.partners, err = a.db.GetPartners()
	if err != nil && !errors.Is(err, storage.ErrPartnersNoFound) {
		return t, err
	}
	// Если партнеров нет, инициализируем пустой список
	if t.partners == nil {
		t.partners = &models.Partners{}
	}

	table := widget.NewTable(
		func() (int, int) {
			return len(*t.partners) + 1, 9 // +1 для заголовков, 6 колонок
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
				case 6:
					label.SetText("Почта")
				case 7:
					label.SetText("Юр. Адрес")
				case 8:
					label.SetText("Скидка")
				}
			} else {
				// Данные партнеров
				if len(*t.partners) > 0 && i.Row-1 < len(*t.partners) {
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
					case 6:
						label.SetText(p.Email)
					case 7:
						label.SetText(p.Address)
					case 8:
						label.SetText(fmt.Sprintf("%d%%", p.Discount))
					}
				}
			}

		},
	)

	// Настройка ширины колонок
	table.SetColumnWidth(0, 50)  // Пустой столбец
	table.SetColumnWidth(1, 200) // Name
	table.SetColumnWidth(2, 120) // Type
	table.SetColumnWidth(3, 150) // Director
	table.SetColumnWidth(4, 120) // Phone
	table.SetColumnWidth(5, 80)  // Rating
	table.SetColumnWidth(6, 150) // Email
	table.SetColumnWidth(7, 200) // Address
	table.SetColumnWidth(8, 150)

	t.table = table
	t.selectPartnerColumn(a)
	t.addPartnerButton(a)
	t.deletePartnerButton(a)

	return t, nil
}

func (t *partnerTable) addPartnerButton(a *App) {
	addButton := widget.NewButton("Добавить Партнера", func() {
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
				t.table.Refresh()
			}
		})
	})
	t.addButton = addButton
}

func (t *partnerTable) deletePartnerButton(a *App) {
	deleteButton := widget.NewButton("Удалить Партнера", func() {
		if t.selectedPartnerID != 0 {
			err := a.db.DeletePartner(t.selectedPartnerID)
			if err != nil {
				dialog.ShowError(err, a.w)
				log.Println(err)
			} else {
				t.partners, err = a.db.GetPartners()
				if err != nil {
					dialog.ShowError(err, a.w)
					log.Println(err)
				}
				t.table.Refresh()
				t.selectedPartnerID = 0 // Сброс выбранного партнера

				// Если партнеров больше нет, выводим сообщение
				if len(*t.partners) == 0 {
					dialog.ShowInformation("Нет данных", "Все партнеры удалены. Добавьте нового партнера.", a.w)
				}
			}
		} else {
			dialog.ShowInformation("Не выбран", "Выберите партнера для удаления", a.w)
		}
	})
	t.deleteButton = deleteButton
}

func (t *partnerTable) selectPartnerColumn(a *App) {
	t.table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && len(*t.partners) > 0 { // Кликабельна вся строка, кроме заголовков
			t.selectedPartnerID = (*t.partners)[id.Row-1].Id
			p := (*t.partners)[id.Row-1]
			if id.Col != 0 {
				showPartnerForm(a.w, p, func(updatedPartner models.Partner) {
					err := a.db.UpdatePartner(updatedPartner) // Используем UpdatePartner вместо AddPartner
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
}

func showPartnerForm(w fyne.Window, p models.Partner, onSave func(models.Partner)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(p.CompanyName)

	typeEntry := widget.NewSelect([]string{"ООО", "ИП", "ОАО", "ПАО", "ЗАО"}, func(s string) {
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
		widget.NewFormItem("Название Компании", nameEntry),
		widget.NewFormItem("Тип компании", typeEntry),
		widget.NewFormItem("Директор", directorEntry),
		widget.NewFormItem("Телефон", phoneEntry),
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Юр. Адрес", addressEntry),
		widget.NewFormItem("Рейтинг", ratingEntry),
	)

	// Убираем кнопку Submit из формы
	form.SubmitText = ""
	form.OnSubmit = nil

	dialog.ShowCustomConfirm("Редактировать партнера", "Сохранить", "Отменить", form, func(b bool) {
		if b {
			err := validateForm(nameEntry.Text, typeEntry.Selected, directorEntry.Text, phoneEntry.Text, emailEntry.Text, addressEntry.Text, ratingEntry.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			rating, _ := strconv.Atoi(ratingEntry.Text)
			p.CompanyName = nameEntry.Text
			p.PartnerType = typeEntry.Selected
			p.Director = directorEntry.Text
			p.Phone = phoneEntry.Text
			p.Email = emailEntry.Text
			p.Address = addressEntry.Text
			p.Rating = rating

			onSave(p)
		}
	}, w)
}

// validateForm проверяет все поля на корректность
func validateForm(companyName, partnerType, director, phone, email, address, rating string) error {
	if companyName == "" {
		return fmt.Errorf("Название компании не может быть пустым")
	}
	if partnerType == "" {
		return fmt.Errorf("Тип компании не может быть пустым")
	}
	if director == "" {
		return fmt.Errorf("Имя директора не может быть пустым")
	}
	if phone == "" {
		return fmt.Errorf("Телефон не может быть пустым")
	}
	if email == "" {
		return fmt.Errorf("Email не может быть пустым")
	}
	//if !strings.Contains(email, "@") {
	//	return fmt.Errorf("Email должен содержать символ @")
	//}
	if address == "" {
		return fmt.Errorf("Юридический адрес не может быть пустым")
	}
	if rating == "" {
		return fmt.Errorf("Рейтинг не может быть пустым")
	}

	ratingValue, err := strconv.Atoi(rating)
	if err != nil {
		return fmt.Errorf("Рейтинг должен быть числом")
	}
	if ratingValue < 0 {
		return fmt.Errorf("Рейтинг должен быть положительным числом")
	}

	return nil
}
