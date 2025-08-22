package videoUpscaler

import "cosmossdk.io/errors"

var (
	ErrIndexTooLong     = errors.Register(ModuleName, 2, "index too long")
	ErrDuplicateAddress = errors.Register(ModuleName, 3, "duplicate address")

	ErrWorkerAlreadyRegistered = errors.Register(ModuleName, 10, "worker already registered")
	ErrWorkerNotAvailable      = errors.Register(ModuleName, 11, "worker cannot subscribe to task")
	ErrWorkerTaskNotAvailable  = errors.Register(ModuleName, 12, "task is already completed")
	ErrWorkerIncorrectStake    = errors.Register(ModuleName, 13, "staked coin is incorrect")

	ErrInvalidVideoUpscalerTask = errors.Register(ModuleName, 20, "invalid video upscaler task")

	ErrInvalidSolution = errors.Register(ModuleName, 30, "proposed solution is invalid")

	ErrInvalidVerification = errors.Register(ModuleName, 40, "verification to solution is invalid")
)
