// components contain layout code that are shared by multiple pages but aren't widely used enough to be defined as
// widgets

package components

import (
	"errors"
	"fmt"
	"image/color"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/ararog/timeago"
	"github.com/crypto-power/cryptopower/app"
	"github.com/crypto-power/cryptopower/libwallet/assets/dcr"
	sharedW "github.com/crypto-power/cryptopower/libwallet/assets/wallet"
	"github.com/crypto-power/cryptopower/libwallet/txhelper"
	libutils "github.com/crypto-power/cryptopower/libwallet/utils"
	"github.com/crypto-power/cryptopower/ui/cryptomaterial"
	"github.com/crypto-power/cryptopower/ui/load"
	"github.com/crypto-power/cryptopower/ui/values"
)

const (
	Uint32Size    = 32 // 32 or 64 ? shifting 32-bit value by 32 bits will always clear it
	MaxInt32      = 1<<(Uint32Size-1) - 1
	WalletsPageID = "Wallets"
)

type (
	C = layout.Context
	D = layout.Dimensions

	TxStatus struct {
		Title string
		Icon  *cryptomaterial.Image

		// tx purchase only
		TicketStatus       string
		Color              color.NRGBA
		ProgressBarColor   color.NRGBA
		ProgressTrackColor color.NRGBA
		Background         color.NRGBA
	}

	// CummulativeWalletsBalance defines total balance for all available wallets.
	CummulativeWalletsBalance struct {
		Total                   sharedW.AssetAmount
		Spendable               sharedW.AssetAmount
		ImmatureReward          sharedW.AssetAmount
		ImmatureStakeGeneration sharedW.AssetAmount
		LockedByTickets         sharedW.AssetAmount
		VotingAuthority         sharedW.AssetAmount
		UnConfirmed             sharedW.AssetAmount
	}

	DexServer struct {
		SavedHosts map[string][]byte
	}
)

// Container is simply a wrapper for the Inset type. Its purpose is to differentiate the use of an inset as a padding or
// margin, making it easier to visualize the structure of a layout when reading UI code.
type Container struct {
	Padding layout.Inset
}

func (c Container) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	return c.Padding.Layout(gtx, w)
}

func UniformPadding(gtx layout.Context, body layout.Widget) layout.Dimensions {
	width := gtx.Constraints.Max.X

	padding := values.MarginPadding24

	if (width - 2*gtx.Dp(padding)) > gtx.Dp(values.AppWidth) {
		paddingValue := float32(width-gtx.Dp(values.AppWidth)) / 4
		padding = unit.Dp(paddingValue)
	}

	return layout.Inset{
		Top:    values.MarginPadding24,
		Right:  padding,
		Bottom: values.MarginPadding24,
		Left:   padding,
	}.Layout(gtx, body)
}

func UniformHorizontalPadding(gtx layout.Context, body layout.Widget) layout.Dimensions {
	width := gtx.Constraints.Max.X

	padding := values.MarginPadding24

	if (width - 2*gtx.Dp(padding)) > gtx.Dp(values.AppWidth) {
		paddingValue := float32(width-gtx.Dp(values.AppWidth)) / 3
		padding = unit.Dp(paddingValue)
	}

	return layout.Inset{
		Right: padding,
		Left:  padding,
	}.Layout(gtx, body)
}

func UniformMobile(gtx layout.Context, isHorizontal, withList bool, body layout.Widget) layout.Dimensions {
	insetRight := values.MarginPadding10
	if withList {
		insetRight = values.MarginPadding0
	}

	insetTop := values.MarginPadding24
	if isHorizontal {
		insetTop = values.MarginPadding0
	}

	return layout.Inset{
		Top:   insetTop,
		Right: insetRight,
		Left:  values.MarginPadding10,
	}.Layout(gtx, body)
}

func TransactionTitleIcon(l *load.Load, wal sharedW.Asset, tx *sharedW.Transaction) *TxStatus {
	var txStatus TxStatus

	switch tx.Direction {
	case txhelper.TxDirectionSent:
		txStatus.Title = values.String(values.StrSent)
		txStatus.Icon = l.Theme.Icons.SendIcon
	case txhelper.TxDirectionReceived:
		txStatus.Title = values.String(values.StrReceived)
		txStatus.Icon = l.Theme.Icons.ReceiveIcon
	default:
		txStatus.Title = values.String(values.StrTransferred)
		txStatus.Icon = l.Theme.Icons.Transferred
	}

	// replace icon for staking tx types
	if wal.TxMatchesFilter(tx, libutils.TxFilterStaking) {
		switch tx.Type {
		case txhelper.TxTypeTicketPurchase:
			{
				if wal.TxMatchesFilter(tx, libutils.TxFilterUnmined) {
					txStatus.Title = values.String(values.StrUmined)
					txStatus.Icon = l.Theme.Icons.TicketUnminedIcon
					txStatus.TicketStatus = dcr.TicketStatusUnmined
					txStatus.Color = l.Theme.Color.LightBlue6
					txStatus.Background = l.Theme.Color.LightBlue
				} else if wal.TxMatchesFilter(tx, libutils.TxFilterImmature) {
					txStatus.Title = values.String(values.StrImmature)
					txStatus.Icon = l.Theme.Icons.TicketImmatureIcon
					txStatus.Color = l.Theme.Color.Yellow
					txStatus.TicketStatus = dcr.TicketStatusImmature
					txStatus.ProgressBarColor = l.Theme.Color.OrangeYellow
					txStatus.ProgressTrackColor = l.Theme.Color.Gray6
					txStatus.Background = l.Theme.Color.Yellow
				} else if wal.TxMatchesFilter(tx, libutils.TxFilterLive) {
					txStatus.Title = values.String(values.StrLive)
					txStatus.Icon = l.Theme.Icons.TicketLiveIcon
					txStatus.Color = l.Theme.Color.Success2
					txStatus.TicketStatus = dcr.TicketStatusLive
					txStatus.ProgressBarColor = l.Theme.Color.Success2
					txStatus.ProgressTrackColor = l.Theme.Color.Success2
					txStatus.Background = l.Theme.Color.Success2
				} else if wal.TxMatchesFilter(tx, libutils.TxFilterExpired) {
					txStatus.Title = values.String(values.StrExpired)
					txStatus.Icon = l.Theme.Icons.TicketExpiredIcon
					txStatus.Color = l.Theme.Color.GrayText2
					txStatus.TicketStatus = dcr.TicketStatusExpired
					txStatus.Background = l.Theme.Color.Gray4
				} else {
					ticketSpender, _ := wal.(*dcr.Asset).TicketSpender(tx.Hash)
					if ticketSpender != nil {
						if ticketSpender.Type == txhelper.TxTypeVote {
							txStatus.Title = values.String(values.StrVoted)
							txStatus.Icon = l.Theme.Icons.TicketVotedIcon
							txStatus.Color = l.Theme.Color.Turquoise700
							txStatus.TicketStatus = dcr.TicketStatusVotedOrRevoked
							txStatus.ProgressBarColor = l.Theme.Color.Turquoise300
							txStatus.ProgressTrackColor = l.Theme.Color.Turquoise100
							txStatus.Background = l.Theme.Color.Success2
						} else {
							txStatus.Title = values.String(values.StrRevoked)
							txStatus.Icon = l.Theme.Icons.TicketRevokedIcon
							txStatus.Color = l.Theme.Color.Orange
							txStatus.TicketStatus = dcr.TicketStatusVotedOrRevoked
							txStatus.ProgressBarColor = l.Theme.Color.Danger
							txStatus.ProgressTrackColor = l.Theme.Color.Orange3
							txStatus.Background = l.Theme.Color.Orange2
						}
					}
				}
			}
		case txhelper.TxTypeVote:
			txStatus.Title = values.String(values.StrVote)
			txStatus.Icon = l.Theme.Icons.TicketVotedIcon
			txStatus.Color = l.Theme.Color.Turquoise700
			txStatus.TicketStatus = dcr.TicketStatusVotedOrRevoked
			txStatus.ProgressBarColor = l.Theme.Color.Turquoise300
			txStatus.ProgressTrackColor = l.Theme.Color.Turquoise100
			txStatus.Background = l.Theme.Color.Success2
		default:
			txStatus.Title = values.String(values.StrRevocation)
			txStatus.Icon = l.Theme.Icons.TicketRevokedIcon
			txStatus.Color = l.Theme.Color.Orange
			txStatus.TicketStatus = dcr.TicketStatusVotedOrRevoked
			txStatus.ProgressBarColor = l.Theme.Color.Danger
			txStatus.ProgressTrackColor = l.Theme.Color.Orange3
			txStatus.Background = l.Theme.Color.Orange2
		}
	} else if tx.Type == txhelper.TxTypeMixed {
		txStatus.Title = values.String(values.StrMixed)
		txStatus.Icon = l.Theme.Icons.MixedTx
	}

	return &txStatus
}

// transactionRow is a single transaction row on the transactions and overview page. It lays out a transaction's
// direction, balance, status. isTxPage determines if the transaction should be drawn using the transactions page layout.
func LayoutTransactionRow(gtx layout.Context, l *load.Load, wal sharedW.Asset, tx *sharedW.Transaction, isTxPage bool) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	if wal == nil {
		return D{}
	}

	txStatus := TransactionTitleIcon(l, wal, tx)
	amount := wal.ToAmount(tx.Amount).String()
	assetIcon := CoinImageBySymbol(l, wal.GetAssetType(), wal.IsWatchingOnlyWallet())
	walName := l.Theme.Label(values.TextSize12, wal.GetWalletName())

	insetLeft := values.MarginPadding16
	if !isTxPage {
		insetLeft = values.MarginPadding8
	}

	return cryptomaterial.LinearLayout{
		Orientation: layout.Horizontal,
		Width:       cryptomaterial.MatchParent,
		Height:      cryptomaterial.WrapContent,
		Alignment:   layout.Middle,
		Padding: layout.Inset{
			Top:    values.MarginPadding16,
			Bottom: values.MarginPadding10,
		},
	}.Layout(gtx,
		layout.Rigid(txStatus.Icon.Layout24dp),
		layout.Rigid(func(gtx C) D {
			return cryptomaterial.LinearLayout{
				Width:       cryptomaterial.WrapContent,
				Height:      cryptomaterial.WrapContent,
				Orientation: layout.Vertical,
				Padding:     layout.Inset{Left: insetLeft},
				Direction:   layout.Center,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					if tx.Type == txhelper.TxTypeRegular {
						amount := wal.ToAmount(tx.Amount).String()
						if tx.Direction == txhelper.TxDirectionSent && !strings.Contains(amount, "-") {
							amount = "-" + amount
						}
						return LayoutBalanceSize(gtx, l, amount, values.TextSize18)
					}

					return cryptomaterial.LinearLayout{
						Width:       cryptomaterial.WrapContent,
						Height:      cryptomaterial.WrapContent,
						Orientation: layout.Horizontal,
						Direction:   layout.W,
						Alignment:   layout.Baseline,
					}.Layout(gtx,
						layout.Rigid(l.Theme.Label(values.TextSize18, txStatus.Title).Layout),
						layout.Rigid(func(gtx C) D {
							if isTxPage {
								return D{}
							}
							return layout.E.Layout(gtx, func(gtx C) D {
								return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, assetIcon.Layout12dp)
									}),
									layout.Rigid(func(gtx C) D {
										return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, walName.Layout)
									}),
								)
							})
						}),
					)
				}),
				layout.Rigid(func(gtx C) D {
					if !isTxPage && tx.Type == txhelper.TxTypeRegular {
						return cryptomaterial.LinearLayout{
							Width:       cryptomaterial.WrapContent,
							Height:      cryptomaterial.WrapContent,
							Orientation: layout.Horizontal,
							Direction:   layout.W,
							Alignment:   layout.Middle,
						}.Layout(gtx,
							layout.Rigid(assetIcon.Layout12dp),
							layout.Rigid(func(gtx C) D {
								return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, walName.Layout)
							}),
						)
					}

					return cryptomaterial.LinearLayout{
						Width:       cryptomaterial.WrapContent,
						Height:      cryptomaterial.WrapContent,
						Orientation: layout.Horizontal,
						Alignment:   layout.Middle,
					}.Layout(gtx,
						layout.Rigid(func(gtx C) D {
							if isTxPage {
								return D{}
							}
							if tx.Type == txhelper.TxTypeMixed {
								return cryptomaterial.LinearLayout{
									Width:       cryptomaterial.WrapContent,
									Height:      cryptomaterial.WrapContent,
									Orientation: layout.Horizontal,
									Direction:   layout.W,
									Alignment:   layout.Middle,
								}.Layout(gtx,
									layout.Rigid(func(gtx C) D {
										// mix denomination
										mixedDenom := wal.ToAmount(tx.MixDenomination).String()
										txt := l.Theme.Label(values.TextSize12, mixedDenom)
										txt.Color = l.Theme.Color.GrayText2
										return txt.Layout(gtx)
									}),
									layout.Rigid(func(gtx C) D {
										// Mixed outputs count
										if tx.MixCount > 1 {
											label := l.Theme.Label(values.TextSize12, fmt.Sprintf("x%d", tx.MixCount))
											label.Color = l.Theme.Color.GrayText2
											return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, label.Layout)
										}
										return D{}
									}),
								)
							}

							if isTxPage {
								return D{}
							}

							walBalTxt := l.Theme.Label(values.TextSize12, amount)
							walBalTxt.Color = l.Theme.Color.GrayText2
							return walBalTxt.Layout(gtx)
						}),
						layout.Rigid(func(gtx C) D {
							if dcrAsset, ok := wal.(*dcr.Asset); ok && !isTxPage {
								if ok, _ := dcrAsset.TicketHasVotedOrRevoked(tx.Hash); ok {
									return layout.Inset{
										Left: values.MarginPadding4,
									}.Layout(gtx, func(gtx C) D {
										ic := cryptomaterial.NewIcon(l.Theme.Icons.ImageBrightness1)
										ic.Color = l.Theme.Color.GrayText2
										return ic.Layout(gtx, values.MarginPadding6)
									})
								}
							}
							return D{}
						}),
						layout.Rigid(func(gtx C) D {
							var ticketSpender *sharedW.Transaction
							if dcrAsset, ok := wal.(*dcr.Asset); ok {
								ticketSpender, _ = dcrAsset.TicketSpender(tx.Hash)
							}

							if ticketSpender == nil || isTxPage {
								return D{}
							}
							amnt := wal.ToAmount(ticketSpender.VoteReward).ToCoin()
							txt := fmt.Sprintf("%.2f", amnt)
							if amnt > 0 {
								txt = fmt.Sprintf("+%.2f", amnt)
							}
							return layout.Inset{Left: values.MarginPadding4}.Layout(gtx, l.Theme.Label(values.TextSize12, txt).Layout)
						}),
					)
				}),
			)
		}),
		layout.Flexed(1, func(gtx C) D {
			txSize := values.TextSize16
			if !isTxPage {
				txSize = values.TextSize12
			}
			status := l.Theme.Label(txSize, values.String(values.StrUnknown))
			txConfirmations := TxConfirmations(wal, tx)
			reqConf := wal.RequiredConfirmations()
			if txConfirmations < 1 {
				status = l.Theme.Label(txSize, values.String(values.StrUnconfirmedTx))
				status.Color = l.Theme.Color.GrayText1
			} else if txConfirmations >= reqConf {
				status.Color = l.Theme.Color.GrayText2
				date := time.Unix(tx.Timestamp, 0).Format("Jan 2, 2006")
				timeSplit := time.Unix(tx.Timestamp, 0).Format("03:04:05 PM")
				status.Text = fmt.Sprintf("%v at %v", date, timeSplit)
			} else {
				status = l.Theme.Label(txSize, values.StringF(values.StrTxStatusPending, txConfirmations, reqConf))
				status.Color = l.Theme.Color.GrayText1
			}

			return layout.E.Layout(gtx, func(gtx C) D {
				return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						voteOrRevocationTx := tx.Type == txhelper.TxTypeVote || tx.Type == txhelper.TxTypeRevocation
						if isTxPage && voteOrRevocationTx {
							title := values.String(values.StrRevoke)
							if tx.Type == txhelper.TxTypeVote {
								title = values.String(values.StrVote)
							}

							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									lbl := l.Theme.Label(values.TextSize16, fmt.Sprintf("%dd to %s", tx.DaysToVoteOrRevoke, title))
									lbl.Color = l.Theme.Color.GrayText2
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx C) D {
									return layout.Inset{
										Right: values.MarginPadding5,
										Left:  values.MarginPadding5,
									}.Layout(gtx, func(gtx C) D {
										ic := cryptomaterial.NewIcon(l.Theme.Icons.ImageBrightness1)
										ic.Color = l.Theme.Color.GrayText2
										return ic.Layout(gtx, values.MarginPadding6)
									})
								}),
							)
						}

						return D{}
					}),
					layout.Rigid(func(gtx C) D {
						if !isTxPage {
							return cryptomaterial.LinearLayout{
								Width:       cryptomaterial.WrapContent,
								Height:      cryptomaterial.WrapContent,
								Orientation: layout.Vertical,
								Alignment:   layout.End,
								Direction:   layout.Center,
							}.Layout(gtx,
								layout.Rigid(func(gtx C) D {
									tx := tx
									if wal.TxMatchesFilter(tx, libutils.TxFilterStaking) {
										durationPrefix := values.String(values.StrVoted)
										if tx.Type == txhelper.TxTypeTicketPurchase {
											if wal.TxMatchesFilter(tx, libutils.TxFilterUnmined) {
												durationPrefix = values.String(values.StrUmined)
											} else if wal.TxMatchesFilter(tx, libutils.TxFilterImmature) {
												durationPrefix = values.String(values.StrImmature)
											} else if wal.TxMatchesFilter(tx, libutils.TxFilterLive) {
												durationPrefix = values.String(values.StrLive)
											} else if wal.TxMatchesFilter(tx, libutils.TxFilterExpired) {
												durationPrefix = values.String(values.StrExpired)
											}
										} else if tx.Type == txhelper.TxTypeRevocation {
											durationPrefix = values.String(values.StrRevoked)
										}

										durationTxt := TimeAgo(tx.Timestamp)
										durationTxt = fmt.Sprintf("%s %s", durationPrefix, durationTxt)
										return l.Theme.Label(values.TextSize12, durationTxt).Layout(gtx)
									}
									return D{}
								}),
								layout.Rigid(status.Layout),
							)
						}
						return D{}
					}),
					layout.Rigid(func(gtx C) D {
						if isTxPage {
							return status.Layout(gtx)
						}
						return D{}
					}),
					layout.Rigid(func(gtx C) D {
						isMixedOrRegular := tx.Type == txhelper.TxTypeMixed || tx.Type == txhelper.TxTypeRegular
						if !isTxPage && !isMixedOrRegular {
							return D{}
						}
						statusIcon := l.Theme.Icons.ConfirmIcon
						if TxConfirmations(wal, tx) < wal.RequiredConfirmations() {
							statusIcon = l.Theme.Icons.PendingIcon
						}

						if isTxPage {
							return layout.Inset{
								Left: values.MarginPadding15,
								Top:  values.MarginPadding5,
							}.Layout(gtx, statusIcon.Layout12dp)
						}

						return layout.Inset{
							Left: values.MarginPadding2,
						}.Layout(gtx, statusIcon.Layout12dp)
					}),
				)
			})
		}),
	)

}

func TxConfirmations(wallet sharedW.Asset, transaction *sharedW.Transaction) int32 {
	if transaction.BlockHeight != -1 {
		return (wallet.GetBestBlockHeight() - transaction.BlockHeight) + 1
	}

	return 0
}

func FormatDateOrTime(timestamp int64) string {
	utcTime := time.Unix(timestamp, 0).UTC()
	currentTime := time.Now().UTC()

	if strconv.Itoa(currentTime.Year()) == strconv.Itoa(utcTime.Year()) && currentTime.Month().String() == utcTime.Month().String() {
		if strconv.Itoa(currentTime.Day()) == strconv.Itoa(utcTime.Day()) {
			if strconv.Itoa(currentTime.Hour()) == strconv.Itoa(utcTime.Hour()) {
				return TimeAgo(timestamp)
			}

			return TimeAgo(timestamp)
		} else if currentTime.Day()-1 == utcTime.Day() {
			yesterday := values.String(values.StrYesterday)
			return yesterday
		}
	}

	t := strings.Split(utcTime.Format(time.UnixDate), " ")
	t2 := t[2]
	year := strconv.Itoa(utcTime.Year())
	if t[2] == "" {
		t2 = t[3]
	}
	return fmt.Sprintf("%s %s, %s", t[1], t2, year)
}

// EndToEndRow layouts out its content on both ends of its horizontal layout.
func EndToEndRow(gtx layout.Context, leftWidget, rightWidget func(C) D) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(leftWidget),
		layout.Flexed(1, func(gtx C) D {
			return layout.E.Layout(gtx, rightWidget)
		}),
	)
}

func TimeAgo(timestamp int64) string {
	timeAgo, _ := timeago.TimeAgoWithTime(time.Now(), time.Unix(timestamp, 0))
	return timeAgo
}

func TruncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}

func GoToURL(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TimeFormat(secs int, long bool) string {
	var val string
	if secs > 86399 {
		val = "d"
		if long {
			val = " " + values.String(values.StrDays)
		}
		days := secs / 86400
		return fmt.Sprintf("%d%s", days, val)
	} else if secs > 3599 {
		val = "h"
		if long {
			val = " " + values.String(values.StrHours)
		}
		hours := secs / 3600
		return fmt.Sprintf("%d%s", hours, val)
	} else if secs > 59 {
		val = "s"
		if long {
			val = " " + values.String(values.StrMinutes)
		}
		mins := secs / 60
		return fmt.Sprintf("%d%s", mins, val)
	}

	val = "s"
	if long {
		val = " " + values.String(values.StrSeconds)
	}
	return fmt.Sprintf("%d %s", secs, val)
}

// TxPageDropDownFields returns the fields for the required drop down with the
// transactions view page. Since maps access of items order is always random
// an array of keys is provided guarrantee the dropdown order will always be
// maintained.
func TxPageDropDownFields(wType libutils.AssetType, tabIndex int) (mapInfo map[string]int32, keysInfo []string) {
	switch {
	case wType == libutils.BTCWalletAsset && tabIndex == 0:
		// BTC Transactions Activities dropdown fields.
		mapInfo = map[string]int32{
			values.String(values.StrAll):      libutils.TxFilterAll,
			values.String(values.StrSent):     libutils.TxFilterSent,
			values.String(values.StrReceived): libutils.TxFilterReceived,
		}
		keysInfo = []string{
			values.String(values.StrAll),
			values.String(values.StrSent),
			values.String(values.StrReceived),
		}
	case wType == libutils.LTCWalletAsset && tabIndex == 0:
		// LTC Transactions Activities dropdown fields.
		mapInfo = map[string]int32{
			values.String(values.StrAll):      libutils.TxFilterAll,
			values.String(values.StrSent):     libutils.TxFilterSent,
			values.String(values.StrReceived): libutils.TxFilterReceived,
		}
		keysInfo = []string{
			values.String(values.StrAll),
			values.String(values.StrSent),
			values.String(values.StrReceived),
		}
	case wType == libutils.DCRWalletAsset && tabIndex == 0:
		// DCR Transactions Activities dropdown fields.
		mapInfo = map[string]int32{
			values.String(values.StrAll):         libutils.TxFilterAllTx,
			values.String(values.StrSent):        libutils.TxFilterSent,
			values.String(values.StrReceived):    libutils.TxFilterReceived,
			values.String(values.StrTransferred): libutils.TxFilterTransferred,
			values.String(values.StrMixed):       libutils.TxFilterMixed,
		}
		keysInfo = []string{
			values.String(values.StrAll),
			values.String(values.StrSent),
			values.String(values.StrReceived),
			values.String(values.StrTransferred),
			values.String(values.StrMixed),
		}
	case wType == libutils.DCRWalletAsset && tabIndex == 1:
		// DCR staking Activities dropdown fields.
		mapInfo = map[string]int32{
			values.String(values.StrAll):        libutils.TxFilterStaking,
			values.String(values.StrVote):       libutils.TxFilterVoted,
			values.String(values.StrRevocation): libutils.TxFilterRevoked,
		}
		keysInfo = []string{
			values.String(values.StrAll),
			values.String(values.StrVote),
			values.String(values.StrRevocation),
		}
	}
	return
}

// CoinImageBySymbol returns image widget for supported asset coins.
func CoinImageBySymbol(l *load.Load, assetType libutils.AssetType, isWatchOnly bool) *cryptomaterial.Image {
	switch assetType.ToStringLower() {
	case libutils.BTCWalletAsset.ToStringLower():
		if isWatchOnly {
			return l.Theme.Icons.BtcWatchOnly
		}
		return l.Theme.Icons.BTC
	case libutils.DCRWalletAsset.ToStringLower():
		if isWatchOnly {
			return l.Theme.Icons.DcrWatchOnly
		}
		return l.Theme.Icons.DCR
	case libutils.LTCWalletAsset.ToStringLower():
		if isWatchOnly {
			return l.Theme.Icons.LtcWatchOnly
		}
		return l.Theme.Icons.LTC
	}
	return nil
}

// GetTicketPurchaseAccount returns the validly set ticket purchase account if it exists.
func GetTicketPurchaseAccount(selectedWallet *dcr.Asset) (acct *sharedW.Account, err error) {
	tbConfig := selectedWallet.AutoTicketsBuyerConfig()

	isPurchaseAccountSet := tbConfig.PurchaseAccount != -1
	isMixerAccountSet := tbConfig.PurchaseAccount == selectedWallet.MixedAccountNumber()
	isSpendUnmixedAllowed := selectedWallet.ReadBoolConfigValueForKey(sharedW.SpendUnmixedFundsKey, false)
	isAccountMixerConfigSet := selectedWallet.ReadBoolConfigValueForKey(sharedW.AccountMixerConfigSet, false)

	if isPurchaseAccountSet {
		acct, err = selectedWallet.GetAccount(tbConfig.PurchaseAccount)

		if isAccountMixerConfigSet && !isSpendUnmixedAllowed && isMixerAccountSet && err == nil {
			// Mixer account is set and spending from unmixed account is blocked.
			return
		} else if isSpendUnmixedAllowed && err == nil {
			// Spending from unmixed account is allowed. Choose the set account whether its mixed or not.
			return
		}
		// invalid account found. Set it to nil
		acct = nil
	}
	return
}

func CalculateMixedAccountBalance(selectedWallet *dcr.Asset) (*CummulativeWalletsBalance, error) {
	if selectedWallet == nil {
		return nil, errors.New("mixed account only supported by DCR asset")
	}

	// ignore the returned because an invalid purchase account was set previously.
	// Proceed to go and select a valid account if one wasn't provided.
	account, _ := GetTicketPurchaseAccount(selectedWallet)

	var err error
	if account == nil {
		// A valid purchase account hasn't been set. Use default mixed account.
		account, err = selectedWallet.GetAccount(selectedWallet.MixedAccountNumber())
		if err != nil {
			return nil, err
		}
	}

	return &CummulativeWalletsBalance{
		Total:                   account.Balance.Total,
		Spendable:               account.Balance.Spendable,
		ImmatureReward:          account.Balance.ImmatureReward,
		ImmatureStakeGeneration: account.Balance.ImmatureStakeGeneration,
		LockedByTickets:         account.Balance.LockedByTickets,
		VotingAuthority:         account.Balance.VotingAuthority,
		UnConfirmed:             account.Balance.UnConfirmed,
	}, nil
}

func CalculateTotalWalletsBalance(l *load.Load) (*CummulativeWalletsBalance, error) {
	var totalBalance, spandableBalance, immatureReward, votingAuthority,
		immatureStakeGeneration, lockedByTickets, unConfirmed int64

	accountsResult, err := l.WL.SelectedWallet.Wallet.GetAccountsRaw()
	if err != nil {
		return nil, err
	}

	for _, account := range accountsResult.Accounts {
		totalBalance += account.Balance.Total.ToInt()
		spandableBalance += account.Balance.Spendable.ToInt()
		immatureReward += account.Balance.ImmatureReward.ToInt()

		if l.WL.SelectedWallet.Wallet.GetAssetType() == libutils.DCRWalletAsset {
			// Fields required only by DCR
			immatureStakeGeneration += account.Balance.ImmatureStakeGeneration.ToInt()
			lockedByTickets += account.Balance.LockedByTickets.ToInt()
			votingAuthority += account.Balance.VotingAuthority.ToInt()
			unConfirmed += account.Balance.UnConfirmed.ToInt()
		}
	}

	toAmount := func(v int64) sharedW.AssetAmount {
		return l.WL.SelectedWallet.Wallet.ToAmount(v)
	}

	cumm := &CummulativeWalletsBalance{
		Total:                   toAmount(totalBalance),
		Spendable:               toAmount(spandableBalance),
		ImmatureReward:          toAmount(immatureReward),
		ImmatureStakeGeneration: toAmount(immatureStakeGeneration),
		LockedByTickets:         toAmount(lockedByTickets),
		VotingAuthority:         toAmount(votingAuthority),
		UnConfirmed:             toAmount(unConfirmed),
	}

	return cumm, nil
}

func calculateTotalAssetsBalance(l *load.Load) (map[libutils.AssetType]int64, error) {
	wallets := l.WL.AssetsManager.AllWallets()
	assetsTotalBalance := make(map[libutils.AssetType]int64)

	for _, wal := range wallets {
		if wal.IsWatchingOnlyWallet() {
			continue
		}

		accountsResult, err := wal.GetAccountsRaw()
		if err != nil {
			return nil, err
		}

		for _, account := range accountsResult.Accounts {
			assetsTotalBalance[wal.GetAssetType()] += account.Balance.Total.ToInt()
		}
	}

	return assetsTotalBalance, nil
}

func CalculateTotalAssetsBalance(l *load.Load) (map[libutils.AssetType]sharedW.AssetAmount, error) {
	balances, err := calculateTotalAssetsBalance(l)
	if err != nil {
		return nil, err
	}

	assetsTotalBalance := make(map[libutils.AssetType]sharedW.AssetAmount)
	for assetType, balance := range balances {
		switch assetType {
		case libutils.BTCWalletAsset:
			assetsTotalBalance[assetType] = l.WL.AssetsManager.AllBTCWallets()[0].ToAmount(balance)
		case libutils.DCRWalletAsset:
			assetsTotalBalance[assetType] = l.WL.AssetsManager.AllDCRWallets()[0].ToAmount(balance)
		case libutils.LTCWalletAsset:
			assetsTotalBalance[assetType] = l.WL.AssetsManager.AllLTCWallets()[0].ToAmount(balance)
		default:
			return nil, fmt.Errorf("Unsupported asset type: %s", assetType)
		}
	}

	return assetsTotalBalance, nil
}

func CalculateAssetsUSDBalance(l *load.Load, assetsTotalBalance map[libutils.AssetType]sharedW.AssetAmount) (map[libutils.AssetType]float64, error) {
	usdBalance := func(bal sharedW.AssetAmount, market string) (float64, error) {
		rate := l.WL.AssetsManager.RateSource.GetTicker(market)
		if rate == nil || rate.LastTradePrice <= 0 {
			return 0, fmt.Errorf("No rate information available")
		}

		return bal.MulF64(rate.LastTradePrice).ToCoin(), nil
	}

	assetsTotalUSDBalance := make(map[libutils.AssetType]float64)
	for assetType, balance := range assetsTotalBalance {
		marketValue, exist := values.AssetExchangeMarketValue[assetType]
		if !exist {
			return nil, fmt.Errorf("Unsupported asset type: %s", assetType)
		}
		usdBal, err := usdBalance(balance, marketValue)
		if err != nil {
			return nil, err
		}
		assetsTotalUSDBalance[assetType] = usdBal
	}

	return assetsTotalUSDBalance, nil
}

// SecondsToDays takes time in seconds and returns its string equivalent in the format ddhhmm.
func SecondsToDays(totalTimeLeft int64) string {
	q, r := divMod(totalTimeLeft, 24*60*60)
	timeLeft := time.Duration(r) * time.Second
	if q > 0 {
		return fmt.Sprintf("%dd%s", q, timeLeft.String())
	}
	return timeLeft.String()
}

// divMod divides a numerator by a denominator and returns its quotient and remainder.
func divMod(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

func BrowserURLWidget(gtx C, l *load.Load, url string, copyRedirect *cryptomaterial.Clickable) D {
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			border := widget.Border{Color: l.Theme.Color.Gray4, CornerRadius: values.MarginPadding10, Width: values.MarginPadding2}
			wrapper := l.Theme.Card()
			wrapper.Color = l.Theme.Color.Gray4
			return border.Layout(gtx, func(gtx C) D {
				return wrapper.Layout(gtx, func(gtx C) D {
					return layout.UniformInset(values.MarginPadding10).Layout(gtx, func(gtx C) D {
						return layout.Flex{}.Layout(gtx,
							layout.Flexed(0.9, l.Theme.Body1(url).Layout),
							layout.Flexed(0.1, func(gtx C) D {
								return layout.E.Layout(gtx, func(gtx C) D {
									if copyRedirect.Clicked() {
										clipboard.WriteOp{Text: url}.Add(gtx.Ops)
										l.Toast.Notify(values.String(values.StrCopied))
									}
									return copyRedirect.Layout(gtx, l.Theme.Icons.CopyIcon.Layout24dp)
								})
							}),
						)
					})
				})
			})
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:  values.MarginPaddingMinus10,
				Left: values.MarginPadding10,
			}.Layout(gtx, func(gtx C) D {
				label := l.Theme.Body2(values.String(values.StrWebURL))
				label.Color = l.Theme.Color.GrayText2
				return label.Layout(gtx)
			})
		}),
	)
}

// IsFetchExchangeRateAPIAllowed returns true if the exchange rate fetch API is
// allowed.
func IsFetchExchangeRateAPIAllowed(wl *load.WalletLoad) bool {
	return wl.AssetsManager.GetCurrencyConversionExchange() != values.DefaultExchangeValue &&
		!wl.AssetsManager.IsPrivacyModeOn()
}

// DisablePageWithOverlay disables the provided page by highlighting a message why
// the page is disabled and adding a background color overlay that blocks any
// page event being triggered.
func DisablePageWithOverlay(l *load.Load, currentPage app.Page, gtx C, txt string, actionButton *cryptomaterial.Button) D {
	return layout.Stack{Alignment: layout.N}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			if currentPage == nil {
				return D{}
			}
			mgtx := gtx.Disabled()
			return currentPage.Layout(mgtx)
		}),
		layout.Stacked(func(gtx C) D {
			overlayColor := l.Theme.Color.Gray3
			overlayColor.A = 220
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
			cryptomaterial.FillMax(gtx, overlayColor, 10)

			lbl := l.Theme.Label(values.TextSize20, txt)
			lbl.Font.Weight = font.SemiBold
			lbl.Color = l.Theme.Color.PageNavText
			return layout.Center.Layout(gtx, func(gtx C) D {
				return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx C) D {
						return layout.Inset{Bottom: values.MarginPadding20}.Layout(gtx.Disabled(), lbl.Layout)
					}),
					layout.Rigid(func(gtx C) D {
						if actionButton != nil {
							actionButton.TextSize = values.TextSize14
							return actionButton.Layout(gtx)
						}
						return D{}
					}),
				)
			})
		}),
	)
}

func WalletHightlighLabel(theme *cryptomaterial.Theme, gtx C, textSize unit.Sp, content string) D {
	indexLabel := theme.Label(textSize, content)
	indexLabel.Color = theme.Color.PageNavText
	indexLabel.Font.Weight = font.Medium
	return cryptomaterial.LinearLayout{
		Width:      cryptomaterial.WrapContent,
		Height:     gtx.Dp(values.MarginPadding22),
		Direction:  layout.Center,
		Background: theme.Color.Gray8,
		Padding: layout.Inset{
			Left:  values.MarginPadding8,
			Right: values.MarginPadding8},
		Margin: layout.Inset{Right: values.MarginPadding8},
		Border: cryptomaterial.Border{Radius: cryptomaterial.Radius(9), Color: theme.Color.Gray3, Width: values.MarginPadding1},
	}.Layout2(gtx, indexLabel.Layout)
}

// InputsNotEmpty checks if all the provided editors have non-empty text.
func InputsNotEmpty(editors ...*widget.Editor) bool {
	for _, e := range editors {
		if e.Text() == "" {
			return false
		}
	}
	return true
}
