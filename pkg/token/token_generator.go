package token

import (
	"github.com/bwmarrin/snowflake"
	"log/slog"
	"time"
)

type TokenGenerator interface {
	GenerateToken() (snowflake.ID, error)
}

type SnowflakeTokenGenerator struct {
	epoch  string
	logger *slog.Logger
}

func NewSnowflakeTokenGenerator(epoch string, log *slog.Logger) TokenGenerator {
	return SnowflakeTokenGenerator{
		epoch:  epoch,
		logger: log,
	}
}

func (s SnowflakeTokenGenerator) GenerateToken() (snowflake.ID, error) {
	s.logger.Debug("Generating token", "epoch", s.epoch)
	epoch, err := time.Parse(time.RFC3339, s.epoch)
	if err != nil {
		s.logger.Error("Failed to parse epoch", "err", err)
		return 0, err
	}

	snowflake.Epoch = epoch.UnixNano() / 1e6 //TODO check:apparently this converst it to ms

	node, err := snowflake.NewNode(1) //TODO this is the machine id, if I get this right this is the instance
	if err != nil {
		s.logger.Error("Failed to create node", "err", err)
		return 0, err
	}

	id := node.Generate()
	s.logger.Debug("Generated token", "id", id)

	return id, nil
}

type FakeTokenGenerator struct {
	GenerateTokenFn func() (snowflake.ID, error)
}

func (f *FakeTokenGenerator) GenerateToken() (snowflake.ID, error) {
	return f.GenerateTokenFn()
}
