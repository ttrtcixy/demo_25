package application

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
)

func (a *App) createMaterialsCalcTab() fyne.CanvasObject {
	// Получаем данные из БД
	products, err := a.db.GetProducts()
	if err != nil {
		return widget.NewLabel("Ошибка загрузки продуктов: " + err.Error())
	}

	materialTypes, err := a.db.GetMaterialTypes()
	if err != nil {
		return widget.NewLabel("Ошибка загрузки типов материалов: " + err.Error())
	}

	// Создаем элементы формы
	productSelect := widget.NewSelect(products, nil)
	materialSelect := widget.NewSelect(materialTypes, nil)
	quantityEntry := widget.NewEntry()
	param1Entry := widget.NewEntry()
	param2Entry := widget.NewEntry()

	// Настройка полей
	quantityEntry.SetPlaceHolder("Количество")
	quantityEntry.Validator = validation.NewRegexp(`^[1-9]\d*$`, "Должно быть целое число > 0")
	param1Entry.SetPlaceHolder("Параметр 1")
	param1Entry.Validator = validation.NewRegexp(`^[0-9]*\.?[0-9]+$`, "Должно быть число > 0")
	param2Entry.SetPlaceHolder("Параметр 2")
	param2Entry.Validator = validation.NewRegexp(`^[0-9]*\.?[0-9]+$`, "Должно быть число > 0")

	// Поле результата
	resultLabel := widget.NewLabel("")
	resultLabel.TextStyle.Bold = true

	// Кнопка расчета
	calculateBtn := widget.NewButton("Рассчитать", func() {
		// Проверка заполнения
		if productSelect.Selected == "" || materialSelect.Selected == "" {
			resultLabel.SetText("Выберите продукт и материал")
			return
		}

		// Парсинг параметров
		quantity, err := strconv.Atoi(quantityEntry.Text)
		if err != nil {
			resultLabel.SetText("Некорректное количество")
			return
		}

		param1, err := strconv.ParseFloat(param1Entry.Text, 64)
		if err != nil {
			resultLabel.SetText("Некорректный параметр 1")
			return
		}

		param2, err := strconv.ParseFloat(param2Entry.Text, 64)
		if err != nil {
			resultLabel.SetText("Некорректный параметр 2")
			return
		}

		// Получаем ID из выбранных значений
		productId := strings.Split(productSelect.Selected, " - ")[0]
		materialId := strings.Split(materialSelect.Selected, " - ")[0]

		// Выполняем расчет
		required, err := a.db.CalculateMaterial(productId, materialId, quantity, param1, param2)
		if err != nil {
			resultLabel.SetText("Ошибка расчета: " + err.Error())
			return
		}

		resultLabel.SetText(fmt.Sprintf("Требуется материала: %d единиц", required))
	})

	// Компоновка интерфейса
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Продукт:", Widget: productSelect},
			{Text: "Материал:", Widget: materialSelect},
			{Text: "Количество:", Widget: quantityEntry},
			{Text: "Параметр 1:", Widget: param1Entry},
			{Text: "Параметр 2:", Widget: param2Entry},
		},
	}

	return container.NewVBox(
		widget.NewLabel("Расчет необходимого материала"),
		widget.NewSeparator(),
		form,
		calculateBtn,
		widget.NewSeparator(),
		resultLabel,
	)
}
