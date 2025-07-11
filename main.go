package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SilhouetteApp struct {
	window       fyne.Window
	folderPath   string
	images       []string
	currentImage int
	threshold    float64
	
	// UI elements
	folderLabel     *widget.Label
	originalImage   *canvas.Image
	silhouetteImage *canvas.Image
	thresholdSlider *widget.Slider
	prevButton      *widget.Button
	nextButton      *widget.Button
	saveButton      *widget.Button
	imageCounter    *widget.Label
}

// Custom theme for space-like appearance
type SpaceTheme struct{}

func (t *SpaceTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{0x0a, 0x0a, 0x1a, 0xff} // Deep space blue
	case theme.ColorNameButton:
		return color.NRGBA{0x2a, 0x1a, 0x4a, 0xff} // Purple nebula
	case theme.ColorNameDisabledButton:
		return color.NRGBA{0x1a, 0x1a, 0x2a, 0xff} // Darker purple
	case theme.ColorNamePrimary:
		return color.NRGBA{0x7a, 0x4a, 0xff, 0xff} // Bright purple
	case theme.ColorNameHover:
		return color.NRGBA{0x4a, 0x2a, 0x7a, 0xff} // Hover purple
	case theme.ColorNameFocus:
		return color.NRGBA{0x9a, 0x6a, 0xff, 0xff} // Focused purple
	case theme.ColorNameForeground:
		return color.NRGBA{0xfa, 0xfa, 0xff, 0xff} // Starlight white
	case theme.ColorNameDisabled:
		return color.NRGBA{0x6a, 0x6a, 0x8a, 0xff} // Dimmed starlight
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{0x8a, 0x8a, 0xaa, 0xff} // Placeholder gray
	case theme.ColorNamePressed:
		return color.NRGBA{0x1a, 0x0a, 0x3a, 0xff} // Pressed dark
	case theme.ColorNameScrollBar:
		return color.NRGBA{0x3a, 0x2a, 0x5a, 0xff} // Scrollbar purple
	case theme.ColorNameShadow:
		return color.NRGBA{0x05, 0x05, 0x0a, 0x80} // Deep shadow
	case theme.ColorNameInputBackground:
		return color.NRGBA{0x1a, 0x1a, 0x2a, 0xff} // Input background
	case theme.ColorNameMenuBackground:
		return color.NRGBA{0x1a, 0x1a, 0x2a, 0xff} // Menu background
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{0x05, 0x05, 0x0a, 0xcc} // Overlay
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *SpaceTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *SpaceTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *SpaceTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNameCaptionText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 20
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 8
	case theme.SizeNameSeparatorThickness:
		return 2
	case theme.SizeNameInputBorder:
		return 2
	}
	return theme.DefaultTheme().Size(name)
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(&SpaceTheme{})
	
	myWindow := myApp.NewWindow("✨ Cosmic Silhouette Generator ✨")
	myWindow.Resize(fyne.NewSize(1200, 800))

	silApp := &SilhouetteApp{
		window:    myWindow,
		threshold: 0.5,
	}

	silApp.setupUI()
	myWindow.ShowAndRun()
}

func (s *SilhouetteApp) setupUI() {
	// Create star background
	starBg := canvas.NewRectangle(color.NRGBA{0x0a, 0x0a, 0x1a, 0xff})
	
	// Title with cosmic styling
	title := widget.NewRichTextFromMarkdown("# ✨ Cosmic Silhouette Generator ✨\n*Transform your images into stellar silhouettes*")
	title.Wrapping = fyne.TextWrapWord
	
	// Folder selection with cosmic styling
	s.folderLabel = widget.NewLabel("🌌 No cosmic folder selected")
	//s.folderLabel.Wrapping = fyne.TextWrapWord
	
	selectFolderBtn := widget.NewButton("🚀 Select Image Folder", s.selectFolder)
	selectFolderBtn.Importance = widget.HighImportance
	
	folderContainer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(selectFolderBtn, s.folderLabel),
		widget.NewSeparator(),
	)
	
	// Cosmic threshold slider
	s.thresholdSlider = widget.NewSlider(0.1, 1.0)
	s.thresholdSlider.Value = s.threshold
	s.thresholdSlider.Step = 0.05
	s.thresholdSlider.OnChanged = s.onThresholdChanged
	
	thresholdLabel := widget.NewRichTextFromMarkdown("## 🌟 Silhouette Intensity")
	strengthLabels := container.NewHBox(
		widget.NewLabel("Subtle ⭐"),
		layout.NewSpacer(),
		widget.NewLabel("⭐ Intense"),
	)
	
	thresholdContainer := container.NewVBox(
		thresholdLabel,
		s.thresholdSlider,
		strengthLabels,
		widget.NewSeparator(),
	)
	
	// Image display with cosmic cards
	s.originalImage = canvas.NewImageFromResource(theme.DocumentIcon())
	s.originalImage.FillMode = canvas.ImageFillContain
	s.originalImage.SetMinSize(fyne.NewSize(450, 350))
	
	s.silhouetteImage = canvas.NewImageFromResource(theme.DocumentIcon())
	s.silhouetteImage.FillMode = canvas.ImageFillContain
	s.silhouetteImage.SetMinSize(fyne.NewSize(450, 350))
	
	// Create glowing borders for images
	originalBorder := canvas.NewRectangle(color.NRGBA{0x7a, 0x4a, 0xff, 0x80})
	silhouetteBorder := canvas.NewRectangle(color.NRGBA{0x7a, 0x4a, 0xff, 0x80})
	
	originalContainer := container.NewStack(
		originalBorder,
		container.NewPadded(s.originalImage),
	)
	
	silhouetteContainer := container.NewStack(
		silhouetteBorder,
		container.NewPadded(s.silhouetteImage),
	)
	
	originalCard := widget.NewCard("🌍 Original Image", "", originalContainer)
	silhouetteCard := widget.NewCard("🌑 Cosmic Silhouette", "", silhouetteContainer)
	
	imageRow := container.NewHBox(originalCard, silhouetteCard)
	
	// Cosmic navigation buttons
	s.prevButton = widget.NewButton("⬅️ Previous Star", s.previousImage)
	s.nextButton = widget.NewButton("Next Star ➡️", s.nextImage)
	s.saveButton = widget.NewButton("💾 Save to Galaxy", s.saveSilhouette)
	s.imageCounter = widget.NewLabel("🌟 0 / 0")
	
	// Style buttons
	s.prevButton.Importance = widget.MediumImportance
	s.nextButton.Importance = widget.MediumImportance
	s.saveButton.Importance = widget.HighImportance
	
	s.prevButton.Disable()
	s.nextButton.Disable()
	s.saveButton.Disable()
	
	// Create a cosmic navigation panel
	navSpacer1 := layout.NewSpacer()
	navSpacer2 := layout.NewSpacer()
	navSpacer3 := layout.NewSpacer()
	
	navigationContainer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			s.prevButton,
			navSpacer1,
			s.imageCounter,
			navSpacer2,
			s.nextButton,
			navSpacer3,
			s.saveButton,
		),
	)
	
	// Status bar
	statusBar := container.NewHBox(
		widget.NewLabel("🌌 Ready to explore the cosmos"),
		layout.NewSpacer(),
		widget.NewLabel("Made with ✨ and Go"),
	)
	
	// Main layout with cosmic styling
	content := container.NewStack(
		starBg,
		container.NewPadded(
			container.NewVBox(
				title,
				folderContainer,
				thresholdContainer,
				imageRow,
				navigationContainer,
				layout.NewSpacer(),
				statusBar,
			),
		),
	)
	
	s.window.SetContent(content)
}

func (s *SilhouetteApp) selectFolder() {
	dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil || folder == nil {
			return
		}
		
		s.folderPath = folder.Path()
		s.folderLabel.SetText(fmt.Sprintf("🌌 %s", filepath.Base(s.folderPath)))
		s.loadImages()
	}, s.window)
}

func (s *SilhouetteApp) loadImages() {
	s.images = []string{}
	
	err := filepath.Walk(s.folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				s.images = append(s.images, path)
			}
		}
		return nil
	})
	
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	
	if len(s.images) == 0 {
		dialog.ShowInformation("🌌 No Stars Found", "No supported images found in the selected cosmic folder.\nSupported formats: JPG, JPEG, PNG", s.window)
		return
	}
	
	s.currentImage = 0
	s.updateUI()
	s.loadCurrentImage()
}

func (s *SilhouetteApp) updateUI() {
	if len(s.images) == 0 {
		s.prevButton.Disable()
		s.nextButton.Disable()
		s.saveButton.Disable()
		s.imageCounter.SetText("🌟 0 / 0")
		return
	}
	
	s.prevButton.Enable()
	s.nextButton.Enable()
	s.saveButton.Enable()
	
	if s.currentImage == 0 {
		s.prevButton.Disable()
	}
	if s.currentImage == len(s.images)-1 {
		s.nextButton.Disable()
	}
	
	filename := filepath.Base(s.images[s.currentImage])
	s.imageCounter.SetText(fmt.Sprintf("🌟 %d / %d - %s", s.currentImage+1, len(s.images), filename))
}

func (s *SilhouetteApp) loadCurrentImage() {
	if len(s.images) == 0 {
		return
	}

	imagePath := s.images[s.currentImage]

	// Load original image
	file, err := os.Open(imagePath)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}
	defer file.Close()

	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	resource := fyne.NewStaticResource(filepath.Base(imagePath), imgData)
	s.originalImage.Resource = resource
	s.originalImage.Refresh()

	imgDecoded, _, err := image.Decode(strings.NewReader(string(imgData)))
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.generateSilhouette(imgDecoded)
}


func (s *SilhouetteApp) generateSilhouette(img image.Image) {
	bounds := img.Bounds()
	silhouette := image.NewRGBA(bounds)
	
	// Fill with white background
	draw.Draw(silhouette, bounds, &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldColor := img.At(x, y)
			r, g, b, a := oldColor.RGBA()
			
			// Convert to grayscale
			gray := float64(r*299+g*587+b*114) / 1000.0 / 65535.0
			
			// Apply threshold with adjustable strength
			adjustedThreshold := 1.0 - s.threshold
			
			if gray < adjustedThreshold && a > 0 {
				// Make it black (cosmic silhouette)
				silhouette.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
	
	imgBytes := new(strings.Builder)
err := png.Encode(imgBytes, silhouette)
if err != nil {
	dialog.ShowError(err, s.window)
	return
}
resource := fyne.NewStaticResource("cosmic_silhouette.png", []byte(imgBytes.String()))
s.silhouetteImage.Resource = resource
s.silhouetteImage.Refresh()

}

func (s *SilhouetteApp) saveImageToFile(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return png.Encode(file, img)
}

func (s *SilhouetteApp) onThresholdChanged(value float64) {
	s.threshold = value
	s.loadCurrentImage() // Regenerate silhouette with new threshold
}

func (s *SilhouetteApp) previousImage() {
	if s.currentImage > 0 {
		s.currentImage--
		s.updateUI()
		s.loadCurrentImage()
	}
}

func (s *SilhouetteApp) nextImage() {
	if s.currentImage < len(s.images)-1 {
		s.currentImage++
		s.updateUI()
		s.loadCurrentImage()
	}
}

func (s *SilhouetteApp) saveSilhouette() {
	if len(s.images) == 0 {
		return
	}
	
	originalPath := s.images[s.currentImage]
	
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()
		
		// Reload and regenerate silhouette
		file, err := os.Open(originalPath)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		defer file.Close()
		
		img, _, err := image.Decode(file)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		
		// Generate silhouette
		bounds := img.Bounds()
		silhouette := image.NewRGBA(bounds)
		draw.Draw(silhouette, bounds, &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)
		
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				oldColor := img.At(x, y)
				r, g, b, a := oldColor.RGBA()
				gray := float64(r*299+g*587+b*114) / 1000.0 / 65535.0
				adjustedThreshold := 1.0 - s.threshold
				
				if gray < adjustedThreshold && a > 0 {
					silhouette.Set(x, y, color.RGBA{0, 0, 0, 255})
				}
			}
		}
		
		// Save to selected location
		err = png.Encode(writer, silhouette)
		if err != nil {
			dialog.ShowError(err, s.window)
		} else {
			dialog.ShowInformation("🌟 Success!", "Your cosmic silhouette has been saved to the galaxy! ✨", s.window)
		}
	}, s.window)
}