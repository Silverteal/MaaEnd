package itemtransfer

import (
	"image"

	maa "github.com/MaaXYZ/maa-framework-go/v3"
	"github.com/rs/zerolog/log"
)

const (
	RepoFirstX  = 161
	RepoFirstY  = 217
	RepoColumns = 8

	BackpackFirstX  = 771
	BackpackFirstY  = 217
	BackpackColumns = 5

	BackpackRows = 7 // subject to change in later versions of endfield? whatever~

	RowsPerPage  = 4
	BoxSize      = 64
	GridInterval = 5
)

const (
	ToolTipCursorOffset = 32
	TooltipRoiScaleX    = 275
	TooltipRoiScaleY    = 130
)

const (
	RepoTitleX = 185
	RepoTitleY = 130
	RepoTitleW = 145
	RepoTitleH = 40
)

const (
	ResetInvViewSwipeTimes = 5
)

type Inventory int

const (
	REPOSITORY Inventory = iota
	BACKPACK
)

func (inv Inventory) String() string {
	switch inv {
	case REPOSITORY:
		return "Repository"
	case BACKPACK:
		return "Backpack"
	default:
		return "Unknown"
	}
}

func (inv Inventory) FirstX() int {
	switch inv {
	case REPOSITORY:
		return RepoFirstX
	case BACKPACK:
		return BackpackFirstX
	default:
		return 0
	}
}
func (inv Inventory) FirstY() int {
	switch inv {
	case REPOSITORY:
		return RepoFirstY
	case BACKPACK:
		return BackpackFirstY
	default:
		return 0
	}
}

func (inv Inventory) Columns() int {
	switch inv {
	case REPOSITORY:
		return RepoColumns
	case BACKPACK:
		return BackpackColumns
	default:
		return 0
	}
}

type LeftClickWithCtrlDown struct{}

func (*LeftClickWithCtrlDown) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	log.Debug().Msg("Pressing Ctrl")
	success := ctx.RunActionDirect(
		maa.NodeActionTypeKeyDown,
		maa.NodeKeyDownParam{
			Key: 17, // Ctrl
		},
		maa.Rect{},
		nil,
	).Success
	if !success {
		log.Error().
			Msg("Failed pressing ctrl")
		return false
	}
	log.Debug().
		Int("x", arg.RecognitionDetail.Box.X()).
		Int("y", arg.RecognitionDetail.Box.Y()).
		Int("W", arg.RecognitionDetail.Box.Width()).
		Int("H", arg.RecognitionDetail.Box.Height()).
		Msg("Pressed Ctrl. Clicking left mouse.")

	success = ctx.RunActionDirect(
		maa.NodeActionTypeClick,
		maa.NodeClickParam{
			Target: maa.NewTargetRect(arg.RecognitionDetail.Box),
		},
		arg.RecognitionDetail.Box,
		arg.RecognitionDetail,
	).Success
	if !success {
		log.Error().
			Msg("Failed clicking")
		return false
	}
	log.Debug().Msg("Clicked left mouse.")

	success = ctx.RunActionDirect(
		maa.NodeActionTypeKeyUp,
		maa.NodeKeyUpParam{
			Key: 17, // Ctrl
		},
		maa.Rect{},
		nil,
	).Success
	if !success {
		log.Error().
			Msg("Failed releasing ctrl")
		return false
	}
	log.Debug().Msg("Released Ctrl.")
	return true
}

func TooltipRoi(inv Inventory, gridRowY, gridColX int) maa.Rect {
	x := inv.FirstX() + gridColX*(BoxSize+GridInterval) + ToolTipCursorOffset
	y := inv.FirstY() + gridRowY*(BoxSize+GridInterval) + ToolTipCursorOffset
	w := TooltipRoiScaleX
	h := TooltipRoiScaleY
	log.Trace().
		Str("inventory", inv.String()).
		Int("grid_row_y", gridRowY).
		Int("grid_col_x", gridColX).
		Int("x", x).Int("y", y).Int("w", w).Int("h", h).
		Msg("Agent Requested a TOOLTIP ROI")
	return maa.Rect{x, y, w, h}
}

func ItemBoxRoi(inv Inventory, gridRowY, gridColX int) maa.Rect {
	x := inv.FirstX() + gridColX*(BoxSize+GridInterval)
	y := inv.FirstY() + gridRowY*(BoxSize+GridInterval)
	w := BoxSize
	h := BoxSize

	log.Trace().
		Str("inventory", inv.String()).
		Int("grid_row_y", gridRowY).
		Int("grid_col_x", gridColX).
		Int("x", x).Int("y", y).Int("w", w).Int("h", h).
		Msg("Agent Requested a BOX ROI")
	return maa.Rect{x, y, w, h}
}

func HoverOnto(ctx *maa.Context, inv Inventory, gridRowY, gridColX int) (success bool) {
	log.Debug().
		Str("inventory", inv.String()).
		Int("grid_row_y", gridRowY).
		Int("grid_col_x", gridColX).
		Msg("Agent Start Hovering onto Item")
	success = ctx.RunActionDirect(
		maa.NodeActionTypeSwipe,
		maa.NodeSwipeParam{
			OnlyHover: true,
		},
		Pointize(TooltipRoi(inv, gridRowY, gridColX)),
		nil,
	).Success
	if success {
		log.Debug().
			Str("inventory", inv.String()).
			Int("grid_row_y", gridRowY).
			Int("grid_col_x", gridColX).
			Msg("Agent Successfully Hovered to Item")
	} else {
		log.Error().
			Str("inventory", inv.String()).
			Int("grid_row_y", gridRowY).
			Int("grid_col_x", gridColX).
			Msg("Agent Failed Hovering to Item")
	}
	return success
}

func MoveAndShot(ctx *maa.Context, inv Inventory, gridRowY, gridColX int) (img image.Image) {
	// Step 1 - Hover to item
	if !HoverOnto(ctx, inv, gridRowY, gridColX) {
		log.Error().
			Str("inventory", inv.String()).
			Int("grid_row_y", gridRowY).
			Int("grid_col_x", gridColX).
			Msg("Failed to hover onto item")
		return nil
	}

	// Step 2 - Make screenshot
	log.Debug().
		Str("inventory", inv.String()).
		Int("grid_row_y", gridRowY).
		Int("grid_col_x", gridColX).
		Msg("Start Capture")
	controller := ctx.GetTasker().GetController()
	controller.PostScreencap().Wait()
	log.Debug().
		Str("inventory", inv.String()).
		Int("grid_row_y", gridRowY).
		Int("grid_col_x", gridColX).
		Msg("Done Capture")
	return controller.CacheImage()
}

func ResetInventoryView(ctx *maa.Context, inv Inventory, inverse bool) (success bool) {
	log.Debug().
		Str("inventory", inv.String()).
		Msg("Agent Requested a Reset to Inventory View")
	params := maa.NodeSwipeParam{}
	RightUpCorner := Pointize(ItemBoxRoi(inv, 0, inv.Columns()-1))
	RightUpCornerOffset := maa.Rect{BoxSize + (GridInterval / 2), (GridInterval / 2)}
	RightDownCorner := Pointize(ItemBoxRoi(inv, RowsPerPage-1, inv.Columns()-1))
	RightDownCornerOffset := maa.Rect{BoxSize + (GridInterval / 2), BoxSize - (GridInterval / 2)}
	if !inverse {
		params.Begin = maa.NewTargetRect(RightUpCorner)
		params.BeginOffset = RightUpCornerOffset
		params.End = []maa.Target{maa.NewTargetRect(RightDownCorner)}
		params.EndOffset = []maa.Rect{RightDownCornerOffset}
	} else {
		params.Begin = maa.NewTargetRect(RightDownCorner)
		params.BeginOffset = RightDownCornerOffset
		params.End = []maa.Target{maa.NewTargetRect(RightUpCorner)}
		params.EndOffset = []maa.Rect{RightUpCornerOffset}
	}
	for range ResetInvViewSwipeTimes {
		success = ctx.RunActionDirect(
			maa.NodeActionTypeSwipe,
			params,
			maa.Rect{},
			nil,
		).Success
		if !success {
			log.Error().
				Str("inventory", inv.String()).
				Msg("Error occurred while swiping, in ResetInventory")
			return false
		}
		log.Trace().
			Str("inventory", inv.String()).
			Msg("ResetInventory Swiped Once")
	}
	log.Debug().
		Str("inventory", inv.String()).
		Msg("Done Reset Inventory View")
	return true
}
