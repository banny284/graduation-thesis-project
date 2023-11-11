package client

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	tmClient "github.com/tendermint/tendermint/rpc/client"
)

type ChainHeight struct {
	ctx          context.Context
	Logger       zerolog.Logger
	rpc          tmClient.Client
	pollInterval time.Duration
	height       int64
	err          error
}

func NewChainHeight(
	ctx context.Context,
	logger zerolog.Logger,
	rpc tmClient.Client,
	pollInterval time.Duration,
) (*ChainHeight, error) {
	if !rpc.IsRunning() {
		err := rpc.Start()
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to start rpc client")
			return nil, err
		}
	}

	ch := &ChainHeight{
		ctx:          ctx,
		Logger:       logger.With().Str("module", "chain_height").Logger(),
		rpc:          rpc,
		height:       0,
		pollInterval: pollInterval,
		err:          nil,
	}

	ch.update()
	go ch.poll()

	return ch, nil
}

func (c *ChainHeight) poll() {
	for {
		time.Sleep(c.pollInterval)
		c.update()
	}
}

func (c *ChainHeight) update() {
	status, err := c.rpc.Status(c.ctx)
	if err == nil {
		if c.height < status.SyncInfo.LatestBlockHeight {
			c.height = status.SyncInfo.LatestBlockHeight
			c.Logger.Info().Int64("height", c.height).Msg("got new chain height")
		} else {
			c.Logger.Debug().
				Int64("new", status.SyncInfo.LatestBlockHeight).
				Int64("current", c.height).
				Msg("ignoring stale chain height")
		}
	} else {
		c.Logger.Warn().Err(err).Msg("failed to get chain height")
	}
	c.err = err
}

func (c *ChainHeight) GetChainHeight() (int64, error) {
	return c.height, c.err
}
