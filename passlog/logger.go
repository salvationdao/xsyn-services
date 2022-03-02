package passlog

import (
	"os"

	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

var PassLog *zerolog.Logger

func New(environment, level string) {
	log := log_helpers.LoggerInitZero(environment, level)
	if environment == "production" || environment == "staging" {
		logPtr := zerolog.New(os.Stdout)
		logPtr = logPtr.With().Caller().Logger()
		log = &logPtr

	}
	log.Info().Msg("zerolog initialised")
	if PassLog != nil {
		panic("passlog already initialised")
	}
	PassLog = log
}