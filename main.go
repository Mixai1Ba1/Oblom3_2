package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var a = 50.0
var b = 150.0

type MenuItemInfo struct {
	Path string
	Item *fyne.MenuItem
}

type TrialResult struct {
	TrialNum     int
	Path         string
	ActualTimeMs int64
	HickTimeMs   float64
}

var (
	allMenuItems []MenuItemInfo
	targetItem   MenuItemInfo
	startTime    time.Time
	trialCount   = 0
	maxTrials    = 10
	results      []TrialResult
)

func main() {
	rand.Seed(time.Now().UnixNano())
	myApp := app.New()
	myWindow := myApp.NewWindow("Тест Хика — Меню")
	myWindow.Resize(fyne.NewSize(900, 600))

	instruction := widget.NewLabel("Добро пожаловать! Нажмите кнопку ниже, чтобы начать тест.")
	output := widget.NewMultiLineEntry()
	output.SetMinRowsVisible(20)
	output.Wrapping = fyne.TextWrapWord
	output.Hide()

	var startButton *widget.Button
	startButton = widget.NewButton("▶️ Начать тест", func() {
		startButton.Hide()
		output.Show()
		instruction.SetText("Тест начался! Следуйте указаниям ниже 👇")
		nextTrial(instruction, output)
	})

	// Главное меню
	fileMenu := fyne.NewMenu("Файл",
		createMenuItem("Создать", output, instruction, "Файл"),
		createMenuItem("Открыть", output, instruction, "Файл"),
		createMenuItem("Сохранить", output, instruction, "Файл"),
		createSubMenu("Сохранить как", []*fyne.MenuItem{
			createMenuItem("Текст RTF", output, instruction, "Файл", "Сохранить как"),
			createSubMenu("Форматы XML", []*fyne.MenuItem{
				createMenuItem("Office Open XML", output, instruction, "Файл", "Сохранить как", "Форматы XML"),
				createMenuItem("OpenDocument", output, instruction, "Файл", "Сохранить как", "Форматы XML"),
			}, output, instruction, "Файл", "Сохранить как"),
			createMenuItem("Простой текст", output, instruction, "Файл", "Сохранить как"),
			createSubMenu("Другие", []*fyne.MenuItem{
				createMenuItem("PDF", output, instruction, "Файл", "Сохранить как", "Другие"),
				createSubMenu("Экспорт в", []*fyne.MenuItem{
					createMenuItem("PNG", output, instruction, "Файл", "Сохранить как", "Другие", "Экспорт в"),
					createMenuItem("JPEG", output, instruction, "Файл", "Сохранить как", "Другие", "Экспорт в"),
				}, output, instruction, "Файл", "Сохранить как", "Другие"),
			}, output, instruction, "Файл", "Сохранить как"),
		}, output, instruction, "Файл"),
		createMenuItem("Печать", output, instruction, "Файл"),
		createMenuItem("Выход", output, instruction, "Файл"),
	)

	editMenu := fyne.NewMenu("Правка",
		createMenuItem("Отменить", output, instruction, "Правка"),
		createMenuItem("Повторить", output, instruction, "Правка"),
		createMenuItem("Копировать", output, instruction, "Правка"),
		createMenuItem("Вставить", output, instruction, "Правка"),
	)

	helpMenu := fyne.NewMenu("Справка",
		createMenuItem("О программе", output, instruction, "Справка"),
	)

	myWindow.SetMainMenu(fyne.NewMainMenu(fileMenu, editMenu, helpMenu))
	myWindow.SetContent(container.NewVBox(instruction, startButton, output))
	myWindow.ShowAndRun()
}

func createMenuItem(label string, output *widget.Entry, instruction *widget.Label, parent ...string) *fyne.MenuItem {
	fullPath := buildFullPath(label, parent)
	item := fyne.NewMenuItem(label, func() {
		fmt.Printf("👉 Нажат пункт: %s\n", fullPath)
		if targetItem.Path == fullPath {
			handleCorrectSelection(fullPath, output, instruction)
		} else {
			showWrongSelection(fullPath, output)
		}
	})
	allMenuItems = append(allMenuItems, MenuItemInfo{Path: fullPath, Item: item})
	return item
}

func createSubMenu(label string, children []*fyne.MenuItem, output *widget.Entry, instruction *widget.Label, parentPath ...string) *fyne.MenuItem {
	fullPath := buildFullPath(label, parentPath)
	subItem := fyne.NewMenuItem(label, nil)
	subItem.ChildMenu = fyne.NewMenu(label, children...)
	for _, child := range children {
		registerNested(child, fullPath, output, instruction)
	}
	return subItem
}

func registerNested(item *fyne.MenuItem, parent string, output *widget.Entry, instruction *widget.Label) {
	fullPath := parent + " → " + item.Label
	if item.ChildMenu != nil {
		for _, child := range item.ChildMenu.Items {
			registerNested(child, fullPath, output, instruction)
		}
	} else {
		item.Action = func() {
			fmt.Printf("👉 Нажат пункт: %s\n", fullPath)
			if targetItem.Path == fullPath {
				handleCorrectSelection(fullPath, output, instruction)
			} else {
				showWrongSelection(fullPath, output)
			}
		}
		allMenuItems = append(allMenuItems, MenuItemInfo{Path: fullPath, Item: item})
	}
}

func nextTrial(instruction *widget.Label, output *widget.Entry) {
	if trialCount >= maxTrials {
		instruction.SetText("✅ Тест завершён!")
		output.SetText(output.Text + "\n🎉 Все 10 попыток завершены!\n")
		showChart(results)
		return
	}

	targetItem = allMenuItems[rand.Intn(len(allMenuItems))]
	startTime = time.Now()
	trialCount++
	text := fmt.Sprintf("Попытка %d: выберите \"%s\"", trialCount, targetItem.Path)
	instruction.SetText(text)
	output.SetText(output.Text + "\n🔔 " + text + "\n")
}

func handleCorrectSelection(label string, output *widget.Entry, instruction *widget.Label) {
	elapsed := time.Since(startTime)
	n := countAlternatives(label)
	hick := hicksLawTime(n)
	results = append(results, TrialResult{
		TrialNum:     trialCount,
		Path:         label,
		ActualTimeMs: elapsed.Milliseconds(),
		HickTimeMs:   hick,
	})
	result := fmt.Sprintf("✅ Вы выбрали: %s\n⏱ Реакция: %d мс\n📊 По Хику: %.2f мс\n",
		label, elapsed.Milliseconds(), hick)
	output.SetText(output.Text + result)
	go func() {
		time.Sleep(1 * time.Second)
		nextTrial(instruction, output)
	}()
}

func showWrongSelection(clickedPath string, output *widget.Entry) {
	msg := fmt.Sprintf("⚠️ Неверный выбор: \"%s\"\n", clickedPath)
	output.SetText(output.Text + msg)

	win := fyne.CurrentApp().Driver().AllWindows()[0]
	dialog.ShowInformation("Неверный пункт", fmt.Sprintf("Вы выбрали: %s\nОжидается: %s", clickedPath, targetItem.Path), win)
}

func countAlternatives(path string) int {
	level := len(strings.Split(path, "→"))
	count := 0
	for _, item := range allMenuItems {
		if len(strings.Split(item.Path, "→")) == level {
			count++
		}
	}
	return count
}

func hicksLawTime(n int) float64 {
	return a + b*math.Log2(float64(n)+1)
}

func buildFullPath(label string, parent []string) string {
	if len(parent) == 0 {
		return label
	}
	return strings.Join(append(parent, label), " → ")
}

func showChart(results []TrialResult) {
	win := fyne.CurrentApp().NewWindow("Результаты по попыткам")
	win.Resize(fyne.NewSize(800, 600))

	var bars []fyne.CanvasObject
	var maxVal int64 = 1

	for _, r := range results {
		if r.ActualTimeMs > maxVal {
			maxVal = r.ActualTimeMs
		}
	}

	for _, r := range results {
		lbl := canvas.NewText(fmt.Sprintf("Попытка %d", r.TrialNum), theme.ForegroundColor())
		lbl.TextSize = 16

		barActual := canvas.NewRectangle(theme.PrimaryColor())
		barHick := canvas.NewRectangle(theme.DisabledColor())

		barActual.SetMinSize(fyne.NewSize(float32(r.ActualTimeMs)*500/float32(maxVal), 20))
		barHick.SetMinSize(fyne.NewSize(float32(r.HickTimeMs)*500/float32(maxVal), 20))

		bars = append(bars, container.NewVBox(
			lbl,
			container.NewHBox(barActual, canvas.NewText(fmt.Sprintf("Факт: %d мс", r.ActualTimeMs), theme.ForegroundColor())),
			container.NewHBox(barHick, canvas.NewText(fmt.Sprintf("Хик: %.0f мс", r.HickTimeMs), theme.ForegroundColor())),
			widget.NewSeparator(),
		))
	}

	scroll := container.NewVScroll(container.NewVBox(bars...))
	win.SetContent(scroll)
	win.Show()
}
