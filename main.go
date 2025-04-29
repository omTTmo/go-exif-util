package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	exiftool "github.com/barasher/go-exiftool"
)

var (
	fileNodes    []string
	list         *widget.List
	currentIndex int
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("AI Prompt EXIF Reader")
	myWindow.Resize(fyne.NewSize(1200, 800))

	output := widget.NewMultiLineEntry()
	output.Wrapping = fyne.TextWrapWord
	output.SetPlaceHolder("Select an image to extract embedded AI prompts...")

	preview := canvas.NewImageFromResource(nil)
	preview.FillMode = canvas.ImageFillContain

	loadFile := func(filePath string) {
		if filePath == "" {
			preview.Image = nil
			preview.Refresh()
			output.SetText("")
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			output.SetText(fmt.Sprintf("Error opening file: %v", err))
			preview.Image = nil
			preview.Refresh()
			return
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			output.SetText(fmt.Sprintf("Error reading file: %v", err))
			preview.Image = nil
			preview.Refresh()
			return
		}

		img, _, imgErr := image.Decode(strings.NewReader(string(buf)))
		if imgErr != nil || img == nil {
			// Try again with bytes.Reader (for binary data)
			imgR := strings.NewReader(string(buf))
			img, _, imgErr = image.Decode(imgR)
		}
		if imgErr == nil && img != nil {
			preview.Image = img
			preview.Refresh()
		} else {
			preview.Image = nil
			preview.Refresh()
		}

		extracted, parsed, err := extractPrompt(filePath)
		if err != nil {
			output.SetText(fmt.Sprintf("Error extracting metadata: %v", err))
			return
		}

		if parsed != "" {
			output.SetText(fmt.Sprintf("RAW:\n%s\n\nPARSED JSON:\n%s", extracted, parsed))
		} else {
			output.SetText(fmt.Sprintf("RAW:\n%s", extracted))
		}
	}

	loadDirectory := func(path string) {
		fileNodes = nil
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			full := filepath.Join(path, name)
			if info, err := e.Info(); err == nil && !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(name))
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
					fileNodes = append(fileNodes, full)
				}
			}
		}
		sort.Strings(fileNodes)
		if list != nil {
			list.Refresh()
			if len(fileNodes) > 0 {
				list.Select(0)
				myWindow.Canvas().Focus(list)
			}
		}
	}

	list = widget.NewList(
		func() int { return len(fileNodes) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			path := fileNodes[id]
			label := obj.(*widget.Label)
			name := filepath.Base(path)
			label.SetText(name)
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(fileNodes) {
			return
		}
		loadFile(fileNodes[id])
	}

	openFolderButton := widget.NewButton("Open Folder", func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				output.SetText("No folder selected.")
				return
			}
			selectedPath := uri.Path()
			loadDirectory(selectedPath)
			// Immediately preview the first image in the selected folder
			for _, p := range fileNodes {
				if info, err := os.Stat(p); err == nil && !info.IsDir() {
					loadFile(p)
					break
				}
			}
		}, myWindow)
		folderDialog.Resize(fyne.NewSize(800, 600))
		folderDialog.Show()
	})

	leftPanel := container.NewBorder(openFolderButton, nil, nil, nil, list)
	centerSplit := container.NewStack(preview)
	rightSplit := container.NewStack(output)

	centerRightSplit := container.NewHSplit(centerSplit, rightSplit)
	centerRightSplit.SetOffset(0.5)

	mainSplit := container.NewHSplit(leftPanel, centerRightSplit)
	mainSplit.SetOffset(0.2)

	myWindow.SetContent(mainSplit)

	showInitialFolderDialog := func() {
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				output.SetText("No folder selected.")
				return
			}
			selectedPath := uri.Path()
			loadDirectory(selectedPath)

			// Auto-select and preview the first image in the selected folder
			if len(fileNodes) > 0 {
				list.Select(0)
				loadFile(fileNodes[0])
			} else {
				preview.Image = nil
				preview.Refresh()
				output.SetText("")
			}
		}, myWindow)
		folderDialog.Resize(fyne.NewSize(800, 600))
		folderDialog.Show()
	}

	myWindow.Show()
	showInitialFolderDialog()
	myApp.Run()
}

func extractPrompt(filePath string) (rawText string, parsedJSON string, err error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize exiftool: %w", err)
	}
	defer et.Close()

	metadata := et.ExtractMetadata(filePath)
	if len(metadata) == 0 {
		return "", "", fmt.Errorf("no metadata found")
	}

	for _, data := range metadata {
		if data.Err != nil {
			continue
		}
		for k, v := range data.Fields {
			if k == "UserComment" || k == "ImageDescription" || k == "XPComment" {
				rawStr := fmt.Sprintf("%v", v)

				if strings.Contains(rawStr, ", ") {
					rawStr = strings.ReplaceAll(rawStr, ", ", ",\n")
				}

				var js interface{}
				if err := json.Unmarshal([]byte(rawStr), &js); err == nil {
					pretty, _ := json.MarshalIndent(js, "", "  ")
					return rawStr, string(pretty), nil
				}

				return rawStr, "", nil
			}
		}
	}

	return "", "", fmt.Errorf("no embedded prompt found")
}
