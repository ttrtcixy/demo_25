package application

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strings"
)

func (a *App) createSalesTab() fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Введите ID партнера или часть имени")
	searchEntry.Resize(fyne.NewSize(400, searchEntry.MinSize().Height))

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	table := widget.NewTable(
		func() (int, int) {
			return 1, 6 // Начинаем с 1 строки (заголовки)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextTruncate
			return container.NewHScroll(label)
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			scrollContainer := o.(*container.Scroll)
			label := scrollContainer.Content.(*widget.Label)

			if i.Row == 0 {
				switch i.Col {
				case 0:
					label.SetText("Продукция")
				case 1:
					label.SetText("Количество")
				case 2:
					label.SetText("Дата продажи")
				case 3:
					label.SetText("Тип продукции")
				case 4:
					label.SetText("Сумма")
				case 5:
					label.SetText("Прибыль")
				}
				label.TextStyle.Bold = true
			} else {

				label.SetText("")
			}
		},
	)

	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 100)
	table.SetColumnWidth(2, 120)
	table.SetColumnWidth(3, 150)
	table.SetColumnWidth(4, 120)
	table.SetColumnWidth(5, 120)

	searchAndDisplay := func() {
		searchTerm := strings.TrimSpace(searchEntry.Text)
		if searchTerm == "" {
			dialog.ShowInformation("Ошибка", "Введите ID или имя партнера", a.w)
			return
		}

		partnerID, partnerName, err := a.db.FindPartner(searchTerm)
		if err != nil {
			dialog.ShowInformation("Не найдено", "Партнер не найден", a.w)
			return
		}

		resultLabel.SetText(fmt.Sprintf("Продажи партнера: %s (ID: %d)", partnerName, partnerID))

		sales, err := a.db.GetPartnerSales(partnerID)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка получения продаж: %v", err), a.w)
			return
		}

		table.Length = func() (int, int) {
			return len(sales) + 1, 6 // +1 для заголовков
		}

		table.UpdateCell = func(i widget.TableCellID, o fyne.CanvasObject) {
			scrollContainer := o.(*container.Scroll)
			label := scrollContainer.Content.(*widget.Label)

			if i.Row == 0 {
				return // Заголовки уже установлены
			}

			if i.Row-1 < len(sales) {
				sale := sales[i.Row-1]
				switch i.Col {
				case 0:
					label.SetText(sale.ProductName)
				case 1:
					label.SetText(fmt.Sprintf("%d", sale.Quantity))
				case 2:
					label.SetText(sale.SaleDate)
				case 3:
					label.SetText(sale.ProductType)
				case 4:
					label.SetText(fmt.Sprintf("%.2f ₽", sale.TotalSum))
				case 5:
					profit := sale.TotalSum * 0.2
					label.SetText(fmt.Sprintf("%.2f ₽", profit))
				}
			}
		}
		table.Refresh()
	}

	searchBtn := widget.NewButton("Поиск", searchAndDisplay)
	searchBtn.Importance = widget.HighImportance
	searchEntry.OnSubmitted = func(_ string) { searchAndDisplay() }

	searchBox := container.NewBorder(
		nil, nil,
		widget.NewLabel("Поиск:"),
		searchBtn,
		searchEntry,
	)

	topPanel := container.NewVBox(
		searchBox,
		resultLabel,
		widget.NewSeparator(),
	)

	tableContainer := container.NewBorder(
		nil, nil, nil, nil,
		container.NewScroll(table),
	)
	tableContainer.Resize(fyne.NewSize(900, 500))

	return container.NewBorder(
		topPanel,
		nil,
		nil,
		nil,
		tableContainer,
	)
}
