package keeper

import (
	"context"
	"errors"
	"log"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/janction/videoUpscaler"
)

var _ videoUpscaler.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) videoUpscaler.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// GetGame defines the handler for the Query/GetGame RPC method.
func (qs queryServer) GetVideoUpscalerTask(ctx context.Context, req *videoUpscaler.QueryGetVideoUpscalerTaskRequest) (*videoUpscaler.QueryGetVideoUpscalerTaskResponse, error) {
	videoUpscalerTask, err := qs.k.VideoUpscalerTasks.Get(ctx, req.Index)
	if err == nil {
		return &videoUpscaler.QueryGetVideoUpscalerTaskResponse{VideoUpscalerTask: &videoUpscalerTask}, nil
	}
	if errors.Is(err, collections.ErrNotFound) {
		return &videoUpscaler.QueryGetVideoUpscalerTaskResponse{VideoUpscalerTask: nil}, nil
	}

	return nil, status.Error(codes.Internal, err.Error())
}

func (qs queryServer) GetVideoUpscalerLogs(ctx context.Context, req *videoUpscaler.QueryGetVideoUpscalerLogsRequest) (*videoUpscaler.QueryGetVideoUpscalerLogsResponse, error) {
	// access database
	var logs []*videoUpscaler.VideoUpscalerLogs_VideoUpscalerLog
	result := qs.k.DB.ReadLogs(req.ThreadId)
	if len(result) == 0 {
		return nil, nil
	}
	for _, val := range result {
		logEntry := videoUpscaler.VideoUpscalerLogs_VideoUpscalerLog{Log: val.Log, Timestamp: val.Timestamp, Severity: videoUpscaler.VideoUpscalerLogs_VideoUpscalerLog_SEVERITY(val.Severity)}
		logs = append(logs, &logEntry)
	}

	return &videoUpscaler.QueryGetVideoUpscalerLogsResponse{VideoUpscalerLogs: &videoUpscaler.VideoUpscalerLogs{ThreadId: req.ThreadId, Logs: logs}}, nil
}

func (qs queryServer) GetPendingVideoUpscalerTasks(ctx context.Context, req *videoUpscaler.QueryGetPendingVideoUpscalerTaskRequest) (*videoUpscaler.QueryGetPendingVideoUpscalerTaskResponse, error) {
	ti, err := qs.k.VideoUpscalerTaskInfo.Get(ctx)

	if err != nil {
		return nil, err
	}
	nextId := ti.NextId

	var result []*videoUpscaler.VideoUpscalerTask
	for i := 0; i < int(nextId); i++ {
		task, err := qs.k.VideoUpscalerTasks.Get(ctx, string(i))
		if err != nil {
			log.Fatalf("unable to retrieve task with id %v. Error: %v", string(i), err.Error())
			continue
		}

		if !task.Completed {
			result = append(result, &task)
		}
	}
	return &videoUpscaler.QueryGetPendingVideoUpscalerTaskResponse{VideoUpscalerTasks: result}, nil
}

func (qs queryServer) GetWorker(ctx context.Context, req *videoUpscaler.QueryGetWorkerRequest) (*videoUpscaler.QueryGetWorkerResponse, error) {
	worker, err := qs.k.Workers.Get(ctx, req.Worker)
	if err != nil {
		return nil, err
	}

	return &videoUpscaler.QueryGetWorkerResponse{Worker: &worker}, nil
}
