package live

import (
	"errors"
)

// errs const
var (
	ErrFailToEnterRoom     = errors.New("fail to enter the room")
	ErrByBlockList         = errors.New("anchor gave you a block list")
	ErrEncryptOrChargeRoom = errors.New("Encrypt or charge rooms")
	ErrNoEntryRoome        = errors.New("no entry room")
)
