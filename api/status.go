package api

import (
	"fmt"
	"github.com/noah-blockchain/noah-go-node/version"
	"github.com/tendermint/tendermint/rpc/core/types"
	"time"
)

type StatusResponse struct {
	NoahVersion       string                   `json:"version"`
	LatestBlockHash   string                   `json:"latest_block_hash"`
	LatestAppHash     string                   `json:"latest_app_hash"`
	LatestBlockHeight int64                    `json:"latest_block_height"`
	LatestBlockTime   time.Time                `json:"latest_block_time"`
	StateHistory      string                   `json:"state_history"`
	TmStatus          *core_types.ResultStatus `json:"tm_status"`
}

func Status() (*StatusResponse, error) {
	result, err := client.Status()
	if err != nil {
		return nil, err
	}

	stateHistory := "off"
	if noahCfg.BaseConfig.KeepStateHistory {
		stateHistory = "on"
	}

	return &StatusResponse{
		NoahVersion:       version.Version,
		LatestBlockHash:   fmt.Sprintf("%X", result.SyncInfo.LatestBlockHash),
		LatestAppHash:     fmt.Sprintf("%X", result.SyncInfo.LatestAppHash),
		LatestBlockHeight: result.SyncInfo.LatestBlockHeight,
		LatestBlockTime:   result.SyncInfo.LatestBlockTime,
		StateHistory:      stateHistory,
		TmStatus:          result,
	}, nil
}
