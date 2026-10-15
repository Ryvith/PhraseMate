package brandicon

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"phrasemate/assets/icons"
)

// Render returns the embedded artwork at the requested pixel size.
// Pre-scaled variants preserve detail at Dock and menu-bar sizes.
func Render(size int) *image.RGBA {
	if size <= 0 {
		return image.NewRGBA(image.Rectangle{})
	}
	name := fmt.Sprintf("icon_%dx%d.png", size, size)
	switch size {
	case 44:
		name = "menu_44.png"
	case 64:
		name = "icon_32x32@2x.png"
	case 1024:
		name = "icon_512x512@2x.png"
	case 16, 32, 128, 256, 512:
	default:
		name = "icon_512x512@2x.png"
	}
	data, err := icons.Files.ReadFile("AppIcon.iconset/" + name)
	if err != nil {
		panic(fmt.Errorf("load embedded icon: %w", err))
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Errorf("decode embedded icon: %w", err))
	}
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if src.Bounds().Dx() == size && src.Bounds().Dy() == size {
		draw.Draw(img, img.Bounds(), src, src.Bounds().Min, draw.Src)
		return img
	}
	// Current callers use the pre-scaled sizes above. Keep arbitrary sizes
	// supported for other platform hosts without adding a resizing dependency.
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, src.At(x*src.Bounds().Dx()/size, y*src.Bounds().Dy()/size))
		}
	}
	return img
}
