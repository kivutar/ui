package app

import (
	"bytes"
	"fmt"
	"image"
	"testing"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

type translucentDamageWidget struct {
	drawTrackingWidget
}

func (w *translucentDamageWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	canvas.DrawRoundRect(geometry.NewRect(20, 20, 100, 35), widget.RGBA(.4, .4, .4, .35), 6)
}

func TestDirtyRepaintClearsPaddingAndGaps(t *testing.T) {
	for _, scale := range []float64{1, 1.25, 4.0 / 3, 1.5, 2.3} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			a := New(WithRenderMode(RenderModeFrameworkManaged), WithTheme(&theme.Theme{}))
			w := a.Window()
			t.Cleanup(w.Close)
			root := &translucentDamageWidget{*newDrawTrackingWidget(geometry.NewRect(0, 0, 180, 100))}
			w.SetRoot(root)
			root.SetRepaintBoundary(false)
			dc := gg.NewContext(180, 100, gg.WithDeviceScale(scale))
			t.Cleanup(func() { _ = dc.Close() })
			canvas := render.NewCanvas(dc, 180, 100)
			w.drawFrameworkManaged(canvas)
			want := dc.Image().(*image.RGBA)

			for _, regions := range [][]geometry.Rect{
				{geometry.NewRect(50, 25, 20, 8)},
				{geometry.NewRect(30, 25, 8, 8), geometry.NewRect(95, 35, 8, 8)},
			} {
				for range 3 {
					w.dirtyTracker.Reset()
					for _, region := range regions {
						w.dirtyTracker.MarkDirty(region)
					}
					if !w.drawFrameworkManaged(canvas) || w.lastWasFullRepaint {
						t.Fatal("expected an incremental repaint")
					}
					if got := dc.Image().(*image.RGBA); !bytes.Equal(got.Pix, want.Pix) {
						t.Fatalf("incremental repaint of %v changed unchanged translucent pixels", regions)
					}
				}
			}
		})
	}
}
