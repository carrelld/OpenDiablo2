package d2interface

import (
	"image"
	"image/color"
)

// Surface is a renderable thing TODO: wtf is it really?
type Surface interface {
	Clear(color color.Color) error
	DrawRect(width, height int, color color.Color)
	DrawLine(x, y int, color color.Color)
	DrawText(format string, params ...interface{})
	GetSize() (width, height int)
	GetDepth() int
	Pop()
	PopN(n int)
	PushColor(color color.Color)
	PushCompositeMode(mode CompositeMode)
	PushFilter(filter Filter)
	PushTranslation(x, y int)
	Render(surface Surface) error
	ReplacePixels(pixels []byte) error
	Screenshot() *image.RGBA
}
