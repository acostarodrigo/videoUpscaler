package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	videoUpscalerv1 "github.com/janction/videoUpscaler/api/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: videoUpscalerv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetVideoUpscalerTask",
					Use:       "get-video-upscaler-task index",
					Short:     "Get the current value of the Video Upscaler task at index",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "index"},
					},
				},
				{
					RpcMethod: "GetPendingVideoUpscalerTasks",
					Use:       "get-pending-video-upscaler-tasks",
					Short:     "Gets the pending video upscaler tasks",
				},
				{
					RpcMethod: "GetWorker",
					Use:       "get-worker worker",
					Short:     "Gets a single worker",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "worker"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: videoUpscalerv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateVideoUpscalerTask",
					Use:       "create-video-upscaler-task [cid] [startFrame] [endFrame] [threads] [reward]",
					Short:     "Creates a new video Upscaler task",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "cid"},
						{ProtoField: "startFrame"},
						{ProtoField: "endFrame"},
						{ProtoField: "threads"},
						{ProtoField: "reward"},
					},
				},
				{
					RpcMethod: "AddWorker",
					Use:       "add-worker [public_ip] [ipfs_id] [stake]--from [workerAddress]",
					Short:     "Registers a new worker that will perform video upscaler tasks",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "public_ip"},
						{ProtoField: "ipfs_id"},
						{ProtoField: "stake"},
					},
				},
				{
					RpcMethod: "SubscribeWorkerToTask",
					Use:       "subscribe-worker-to-task [address] [taskId] [threadId] --from [workerAddress]",
					Short:     "Subscribes an existing enabled worker to perform work in the specified task",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
					},
				},
				{
					RpcMethod: "ProposeSolution",
					Use:       "propose-solution [taskId] [threadId] [publicKey] [signatures] --from [workerAddress]",
					Short:     "Proposes a solution to a thread.",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "public_key"},
						{ProtoField: "signatures", Varargs: true},
					},
				},
				{
					RpcMethod: "SubmitSolution",
					Use:       "submit-solution [taskId] [threadId] [cid] [average_render_seconds] --from [workerAddress]",
					Short:     "Submits the cid of the directory with all the uploaded frames.",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "dir", Varargs: false},
						{ProtoField: "average_render_seconds"},
					},
				},
				{
					RpcMethod: "SubmitValidation",
					Use:       "submit-validation [taskId] [threadId] [publicKey] [signatures] --from [workerAddress]",
					Short:     "Submit a validation to a proposed solution",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "public_key"},
						{ProtoField: "signatures", Varargs: true},
					},
				},
				{
					RpcMethod: "RevealSolution",
					Use:       "reveal-solution [taskId] [threadId] [solution] --from [workerAddress]",
					Short:     "Reveals the CiDs of the solution",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "frames", Varargs: true},
					},
				},
			},
		},
	}
}
