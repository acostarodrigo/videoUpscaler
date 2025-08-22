package videoUpscaler

import "cosmossdk.io/collections"

const ModuleName = "videoUpscaler"

var (
	ParamsKey                    = collections.NewPrefix("Params")
	VideoUpscalerTaskKey         = collections.NewPrefix("videoUpscalerTaskList/value/")
	WorkerKey                    = collections.NewPrefix("Worker")
	TaskInfoKey                  = collections.NewPrefix(0)
	PendingVideoUpscalerTasksKey = collections.NewPrefix(1)
)
