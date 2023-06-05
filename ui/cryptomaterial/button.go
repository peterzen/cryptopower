// SPDX-License-Identifier: Unlicense OR MIT

package cryptomaterial

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"gitlab.com/cryptopower/cryptopower/ui/values"
)

type Button struct {
	material.ButtonStyle
	th                 *Theme
	label              Label
	clickable          *widget.Clickable
	isEnabled          bool
	disabledBackground color.NRGBA
	disabledTextColor  color.NRGBA
	HighlightColor     color.NRGBA

	Margin layout.Inset
}

type ButtonLayout struct {
	material.ButtonLayoutStyle
}

// IconButtonStyle is similar to material.IconButtonStyle but excluding
// the color fields. This ensures that the IconButton colors are only
// set using the IconButton.colorStyle field.
type IconButtonStyle struct {
	Icon   *widget.Icon
	Size   unit.Dp
	Inset  layout.Inset
	Button *widget.Clickable
}

type IconButton struct {
	IconButtonStyle
	colorStyle *values.ColorStyle
}

func (t *Theme) Button(txt string) Button {
	clickable := new(widget.Clickable)
	buttonStyle := material.Button(t.Base, clickable, txt)
	buttonStyle.TextSize = values.TextSize16
	buttonStyle.Background = t.Color.Primary
	buttonStyle.CornerRadius = values.MarginPadding8
	buttonStyle.Inset = layout.Inset{
		Top:    values.MarginPadding10,
		Right:  values.MarginPadding16,
		Bottom: values.MarginPadding10,
		Left:   values.MarginPadding16,
	}

	return Button{
		th:             t,
		ButtonStyle:    buttonStyle,
		label:          t.Label(values.TextSize16, txt),
		clickable:      clickable,
		HighlightColor: t.Color.PrimaryHighlight,
		isEnabled:      true,
	}
}

func (t *Theme) OutlineButton(txt string) Button {
	btn := t.Button(txt)
	btn.Background = color.NRGBA{}
	btn.Color = t.Color.Primary
	btn.HighlightColor = GenHighlightColor(btn.Color)
	btn.disabledBackground = color.NRGBA{}
	btn.disabledTextColor = t.Color.Gray3

	return btn
}

// DangerButton a button with the background set to theme.Danger
func (t *Theme) DangerButton(text string) Button {
	btn := t.Button(text)
	btn.Background = t.Color.Danger
	btn.HighlightColor = GenHighlightColor(btn.Background)
	return btn
}

func GenHighlightColor(c color.NRGBA) color.NRGBA {
	// 127 is transluscent level
	return color.NRGBA{c.R, c.G, c.B, uint8(127)}
}

func (t *Theme) ButtonLayout() ButtonLayout {
	return ButtonLayout{material.ButtonLayout(t.Base, new(widget.Clickable))}
}

func (t *Theme) IconButton(icon *widget.Icon) IconButton {
	return IconButton{
		IconButtonStyle{
			Icon:   icon,
			Button: new(widget.Clickable),
			Size:   unit.Dp(24),
			Inset:  layout.UniformInset(unit.Dp(12)),
		},
		t.Styles.IconButtonColorStyle,
	}
}

func (t *Theme) IconButtonWithStyle(ibs IconButtonStyle, colorStyle *values.ColorStyle) IconButton {
	return IconButton{
		ibs,
		colorStyle,
	}
}

func (b *Button) SetClickable(clickable *widget.Clickable) {
	b.clickable = clickable
}

func (b *Button) SetEnabled(enabled bool) {
	b.isEnabled = enabled
}

func (b *Button) setDisabledColors() {
	b.disabledBackground = b.th.Color.Gray3
	b.disabledTextColor = b.th.Color.Surface
}

func (b *Button) Enabled() bool {
	return b.isEnabled
}

func (b Button) Clicked() bool {
	return b.clickable.Clicked()
}

func (b Button) Hovered() bool {
	return b.clickable.Hovered()
}

func (b Button) Click() {
	b.clickable.Click()
}

func (b *Button) Layout(gtx layout.Context) layout.Dimensions {
	wdg := func(gtx layout.Context) layout.Dimensions {
		return b.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			textColor := b.Color
			if !b.Enabled() {
				b.setDisabledColors()
				textColor = b.disabledTextColor
			}

			b.label.Text = b.Text
			b.label.Font = b.Font
			b.label.Alignment = text.Middle
			b.label.TextSize = b.TextSize
			b.label.Color = textColor
			return b.label.Layout(gtx)
		})
	}

	return b.Margin.Layout(gtx, func(gtx C) D {
		return b.buttonStyleLayout(gtx, wdg)
	})
}

func (b Button) buttonStyleLayout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	min := gtx.Constraints.Min
	return layout.Stack{Alignment: layout.Center}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(b.CornerRadius)
			defer clip.UniformRRect(image.Rectangle{Max: image.Point{
				X: gtx.Constraints.Min.X,
				Y: gtx.Constraints.Min.Y,
			}}, rr).Push(gtx.Ops).Pop()

			background := b.Background
			if !b.Enabled() {
				b.setDisabledColors()
				background = b.disabledBackground
			} else if b.clickable.Hovered() {
				background = Hovered(b.HighlightColor)
			}

			paint.Fill(gtx.Ops, background)
			for _, c := range b.clickable.History() {
				drawInk(gtx, c, b.HighlightColor)
			}

			return layout.Dimensions{Size: gtx.Constraints.Min}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = min
			return layout.Center.Layout(gtx, w)
		}),
		layout.Expanded(func(gtx C) D {
			if !b.Enabled() {
				return D{}
			}

			return b.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				semantic.Button.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			})
		}),
	)
}

func (bl ButtonLayout) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return bl.ButtonLayoutStyle.Layout(gtx, w)
}

// TODO: Test to ensure this works!
// TODO: Doesn't work, if ib.colorStyle was nil before this method is called,
// it is temporarily changed but when ib.Layout is called, it returns to nil.
func (ib IconButton) ChangeColorStyle(colorStyle *values.ColorStyle) {
	// ib.colorStyle = colorStyle ? TODO SA4005: ineffective assignment to field IconButton.colorStyle lint error
}

func (ib IconButton) Layout(gtx layout.Context) layout.Dimensions {
	ibs := material.IconButtonStyle{
		Background: ib.colorStyle.Background,
		Color:      ib.colorStyle.Foreground,
		Icon:       ib.Icon,
		Size:       ib.Size,
		Inset:      ib.Inset,
		Button:     ib.Button,
	}
	return ibs.Layout(gtx)
}

type TextAndIconButton struct {
	theme           *Theme
	Button          *widget.Clickable
	icon            *Icon
	text            string
	Color           color.NRGBA
	BackgroundColor color.NRGBA
}

func (t *Theme) TextAndIconButton(text string, icon *widget.Icon) TextAndIconButton {
	return TextAndIconButton{
		theme:           t,
		Button:          new(widget.Clickable),
		icon:            NewIcon(icon),
		text:            text,
		Color:           t.Color.Surface,
		BackgroundColor: t.Color.Primary,
	}
}

func (b TextAndIconButton) Layout(gtx layout.Context) layout.Dimensions {
	btnLayout := material.ButtonLayout(b.theme.Base, b.Button)
	btnLayout.Background = b.BackgroundColor

	return btnLayout.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx C) D {
			iconAndLabel := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}
			textIconSpacer := unit.Dp(5)

			layIcon := layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: textIconSpacer}.Layout(gtx, func(gtx C) D {
					var d D
					size := gtx.Dp(unit.Dp(46)) - 2*gtx.Dp(unit.Dp(16))
					b.icon.Color = b.Color
					b.icon.Layout(gtx, unit.Dp(14))
					d = layout.Dimensions{
						Size: image.Point{X: size, Y: size},
					}
					return d
				})
			})

			layLabel := layout.Rigid(func(gtx C) D {
				return layout.Inset{Left: textIconSpacer}.Layout(gtx, func(gtx C) D {
					l := material.Label(b.theme.Base, unit.Sp(14), b.text)
					l.Color = b.Color
					return l.Layout(gtx)
				})
			})

			return iconAndLabel.Layout(gtx, layLabel, layIcon)
		})
	})
}
