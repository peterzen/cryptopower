package cryptomaterial

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

type Outline struct {
	BorderColor color.NRGBA
	Weight      int
}

func (t *Theme) Outline() Outline {
	return Outline{
		BorderColor: t.Color.Primary,
		Weight:      2,
	}
}

func (o Outline) Layout(gtx layout.Context, w layout.Widget) D {
	var minHeight int

	dims := layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			borderRadius := gtx.Dp(unit.Dp(4))
			defer clip.RRect{
				Rect: image.Rectangle{Max: image.Point{
					X: gtx.Constraints.Min.X,
					Y: gtx.Constraints.Min.Y,
				}},
				NE: borderRadius, NW: borderRadius, SE: borderRadius, SW: borderRadius,
			}.Push(gtx.Ops).Pop()
			fill(gtx, o.BorderColor)
			minHeight = gtx.Constraints.Min.Y

			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.UniformInset(unit.Dp(1)).Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.Y = minHeight - o.Weight
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					defer clip.RRect{
						Rect: image.Rectangle{Max: image.Point{
							X: gtx.Constraints.Min.X,
							Y: gtx.Constraints.Min.Y,
						}},
						NE: borderRadius, NW: borderRadius, SE: borderRadius, SW: borderRadius,
					}.Push(gtx.Ops).Pop()
					return fill(gtx, rgb(0xffffff))
				})
			})
		}),
		layout.Stacked(func(gtx C) D {
			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return w(gtx)
				})
			})
		}),
	)

	return dims
}
