package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/planetdecred/dcrlibwallet"
	"github.com/planetdecred/godcr/ui/decredmaterial"
	"github.com/planetdecred/godcr/ui/values"
	"github.com/planetdecred/godcr/wallet"
)

const PageWalletSettings = "WalletSettings"

type walletSettingsPage struct {
	theme      *decredmaterial.Theme
	common     *pageCommon
	walletInfo *wallet.MultiWalletInfo
	wal        *wallet.Wallet

	changePass, rescan, deleteWallet *widget.Clickable

	notificationW *widget.Bool
	errorReceiver chan error

	chevronRightIcon *widget.Icon
	backButton       decredmaterial.IconButton
}

func WalletSettingsPage(common *pageCommon) Page {
	pg := &walletSettingsPage{
		theme:         common.theme,
		common:        common,
		walletInfo:    common.info,
		wal:           common.wallet,
		notificationW: new(widget.Bool),
		errorReceiver: make(chan error),

		changePass:   new(widget.Clickable),
		rescan:       new(widget.Clickable),
		deleteWallet: new(widget.Clickable),

		chevronRightIcon: common.icons.chevronRight,
	}

	pg.chevronRightIcon.Color = pg.theme.Color.LightGray
	pg.backButton, _ = common.SubPageHeaderButtons()

	return pg
}

func (pg *walletSettingsPage) OnResume() {

}

func (pg *walletSettingsPage) Layout(gtx layout.Context) layout.Dimensions {
	common := pg.common

	beep := pg.wal.ReadBoolConfigValueForKey(dcrlibwallet.BeepNewBlocksConfigKey)
	pg.notificationW.Value = false
	if beep {
		pg.notificationW.Value = true
	}

	body := func(gtx C) D {
		page := SubPage{
			title:      values.String(values.StrSettings),
			walletName: common.info.Wallets[*common.selectedWallet].Name,
			backButton: pg.backButton,
			back: func() {
				common.changePage(PageWallet)
			},
			body: func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if !common.info.Wallets[*common.selectedWallet].IsWatchingOnly {
							return pg.changePassphrase()(gtx)
						}
						return layout.Dimensions{}
					}),
					layout.Rigid(pg.notification()),
					layout.Rigid(pg.debug()),
					layout.Rigid(pg.dangerZone()),
				)
			},
		}
		return common.SubPageLayout(gtx, page)
	}
	return common.UniformPadding(gtx, body)
}

func (pg *walletSettingsPage) changePassphrase() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrSpendingPassword), pg.changePass, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrChangeSpendingPass))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.chevronRightIcon.Layout(gtx, values.MarginPadding20)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) notification() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrNotifications), nil, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrBeepForNewBlocks))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.theme.Switch(pg.notificationW).Layout(gtx)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) debug() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrDebug), pg.rescan, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrRescanBlockchain))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.chevronRightIcon.Layout(gtx, values.MarginPadding20)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) dangerZone() layout.Widget {
	return func(gtx C) D {
		return pg.pageSections(gtx, values.String(values.StrDangerZone), pg.deleteWallet, func(gtx C) D {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(pg.bottomSectionLabel(values.String(values.StrRemoveWallet))),
				layout.Flexed(1, func(gtx C) D {
					return layout.E.Layout(gtx, func(gtx C) D {
						return pg.chevronRightIcon.Layout(gtx, values.MarginPadding20)
					})
				}),
			)
		})
	}
}

func (pg *walletSettingsPage) pageSections(gtx layout.Context, title string, clickable *widget.Clickable, body layout.Widget) layout.Dimensions {
	dims := func(gtx layout.Context, title string, body layout.Widget) D {
		return layout.UniformInset(values.MarginPadding15).Layout(gtx, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					txt := pg.theme.Body2(title)
					txt.Color = pg.theme.Color.Gray
					return txt.Layout(gtx)
				}),
				layout.Rigid(body),
			)
		})
	}

	return layout.Inset{Bottom: values.MarginPadding10}.Layout(gtx, func(gtx C) D {
		return pg.theme.Card().Layout(gtx, func(gtx C) D {
			if clickable == nil {
				return dims(gtx, title, body)
			}
			return decredmaterial.Clickable(gtx, clickable, func(gtx C) D {
				return dims(gtx, title, body)
			})
		})
	})
}

func (pg *walletSettingsPage) bottomSectionLabel(title string) layout.Widget {
	return func(gtx C) D {
		return pg.theme.Body1(title).Layout(gtx)
	}
}

func (pg *walletSettingsPage) handle() {
	common := pg.common
	for pg.changePass.Clicked() {
		walletID := pg.walletInfo.Wallets[*common.selectedWallet].ID

		newPasswordModal(common).
			title(values.String(values.StrChangeSpendingPass)).
			hint("Current spending password").
			negativeButton(values.String(values.StrCancel), func() {}).
			positiveButton(values.String(values.StrConfirm), func(password string, pm *passwordModal) bool {
				go func() {
					wal := pg.wal.GetMultiWallet().WalletWithID(walletID)
					err := wal.UnlockWallet([]byte(password))
					if err != nil {
						pm.setError(err.Error())
						pm.setLoading(false)
						return
					}
					wal.LockWallet()
					pm.Dismiss()

					// change password
					newCreatePasswordModal(common).
						title(values.String(values.StrChangeSpendingPass)).
						enableName(false).
						passwordHint("New spending password").
						confirmPasswordHint("Confirm new spending password").
						passwordCreated(func(walletName, newPassword string, m *createPasswordModal) bool {
							go func() {
								err := pg.wal.GetMultiWallet().ChangePrivatePassphraseForWallet(walletID, []byte(password),
									[]byte(newPassword), dcrlibwallet.PassphraseTypePass)
								if err != nil {
									m.setError(err.Error())
									m.setLoading(false)
									return
								}
								m.Dismiss()
							}()
							return false
						}).Show()

				}()

				return false
			}).Show()
		break
	}

	for pg.rescan.Clicked() {
		walletID := pg.walletInfo.Wallets[*common.selectedWallet].ID
		go func() {
			info := newInfoModal(common).
				title(values.String(values.StrRescanBlockchain)).
				body("Rescanning may help resolve some balance errors. This will take some time, as it scans the entire"+
					" blockchain for transactions").
				negativeButton(values.String(values.StrCancel), func() {}).
				positiveButton(values.String(values.StrRescan), func() {
					err := pg.wal.RescanBlocks(walletID)
					if err != nil {
						if err.Error() == dcrlibwallet.ErrNotConnected {
							common.notify(values.String(values.StrNotConnected), false)
							return
						}
						common.notify(err.Error(), false)
						return
					}
					msg := values.String(values.StrRescanProgressNotification)
					common.notify(msg, true)
				})

			common.showModal(info)
		}()
		break
	}

	if pg.notificationW.Changed() {
		pg.wal.SaveConfigValueForKey(dcrlibwallet.BeepNewBlocksConfigKey, pg.notificationW.Value)
	}

	for pg.deleteWallet.Clicked() {
		newInfoModal(common).
			title(values.String(values.StrRemoveWallet)).
			body("Make sure to have the seed phrase backed up before removing the wallet").
			negativeButton(values.String(values.StrCancel), func() {}).
			positiveButton(values.String(values.StrRemove), func() {
				walletID := pg.walletInfo.Wallets[*common.selectedWallet].ID

				newPasswordModal(common).
					title(values.String(values.StrConfirmToRemove)).
					negativeButton(values.String(values.StrCancel), func() {}).
					positiveButtonStyle(common.theme.Color.Surface, common.theme.Color.Danger).
					positiveButton(values.String(values.StrConfirm), func(password string, pm *passwordModal) bool {
						pg.wal.DeleteWallet(walletID, []byte(password), pg.errorReceiver)
						return true
					}).Show()

			}).Show()
		break
	}

	select {
	case err := <-pg.errorReceiver:
		if err.Error() == dcrlibwallet.ErrInvalidPassphrase {
			e := values.String(values.StrInvalidPassphrase)
			common.notify(e, false)
		}
	default:
	}
}

func (pg *walletSettingsPage) onClose() {}
