package keeper

import (
	"context"

	"github.com/janction/videoUpscaler"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *videoUpscaler.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	if err := k.VideoUpscalerTaskInfo.Set(ctx, data.VideoUpscalerTaskInfo); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*videoUpscaler.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &videoUpscaler.GenesisState{
		Params:                params,
		VideoUpscalerTaskList: []videoUpscaler.IndexedVideoUpscalerTask{},
		VideoUpscalerTaskInfo: videoUpscaler.VideoUpscalerTaskInfo{NextId: 1},
	}, nil
}
