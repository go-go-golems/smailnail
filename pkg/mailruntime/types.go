package mailruntime

import "github.com/emersion/go-imap/v2"

// UID is an alias for imap.UID.
type UID = imap.UID

// Flag is an alias for imap.Flag.
type Flag = imap.Flag

// StoreFlagsOp is an alias for imap.StoreFlagsOp.
type StoreFlagsOp = imap.StoreFlagsOp

const (
	FlagSeen     = imap.FlagSeen
	FlagAnswered = imap.FlagAnswered
	FlagFlagged  = imap.FlagFlagged
	FlagDeleted  = imap.FlagDeleted
	FlagDraft    = imap.FlagDraft
)

const (
	StoreFlagsAdd = imap.StoreFlagsAdd
	StoreFlagsDel = imap.StoreFlagsDel
	StoreFlagsSet = imap.StoreFlagsSet
)
