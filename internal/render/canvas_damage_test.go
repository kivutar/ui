package render

import (
	"fmt"
	"image"
	"testing"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestCanvasPartialRepaintPreservesTranslucentEdges(t *testing.T) {
	for _, scale := range []float64{0.75, 1, 1.25, 4.0 / 3, 1.5, 1.75, 2, 2.3} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			dc := gg.NewContext(180, 100, gg.WithDeviceScale(scale))
			t.Cleanup(func() { _ = dc.Close() })
			canvas := NewCanvas(dc, 180, 100)
			paint := func() {
				canvas.DrawRoundRect(geometry.NewRect(20, 22, 80, 22), widget.RGBA(.4, .4, .4, .15), 6)
				canvas.DrawRoundRect(geometry.NewRect(20, 21, 80, 22), widget.RGBA(.4, .4, .4, .35), 6)
			}
			paint()
			want := dc.Image().(*image.RGBA)
			// Repaint unchanged content repeatedly, across both the shape's
			// interior and its antialiasing fringe. No pixel should change.
			for _, damage := range []geometry.Rect{
				geometry.NewRect(65.5, 19.5, 28, 18),
				geometry.NewRect(65.5, 41, 28, 18),
				geometry.NewRect(65.5, 43, 28, 18),
				geometry.NewRect(65.5, 43.5, 28, 18),
				geometry.NewRect(19.5, 25.5, 12, 13),
				geometry.NewRect(99.5, 25.5, 12, 13),
				geometry.NewRect(65.5, 25.5, 0, 8),
				geometry.NewRect(65.5, 25.5, 8, 0),
			} {
				for range 3 {
					canvas.FillRectDirect(damage, widget.ColorTransparent)
					canvas.PushClip(damage)
					paint()
					canvas.PopClip()
					got := dc.Image().(*image.RGBA)
					// Clipping can change the rasterizer's final AA rounding by
					// one byte. Larger differences or repeated accumulation are
					// repaint errors.
					for i, value := range got.Pix {
						delta := int(value) - int(want.Pix[i])
						if delta < -1 || delta > 1 {
							t.Fatalf("damage %v changed pixel (%d,%d) channel %d from %d to %d", damage, (i%got.Stride)/4, i/got.Stride, i%4, want.Pix[i], value)
						}
					}
				}
			}
		})
	}
}
