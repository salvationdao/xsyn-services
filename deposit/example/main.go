package main

import (

	// for demo

	"os"
	"passport/deposit"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()
	l := deposit.NewETHERC20Listener()
	xferCh := l.Listen()
	for msg := range xferCh {
		log.Info().
			Str("chain", msg.Chain).
			Int("confirmations", msg.Confirmations).
			Str("contract", msg.Contract).
			Str("txid", msg.TXID).
			Str("from", msg.From).
			Str("to", msg.To).
			Str("value", msg.Value.String()).Msg("transfer")
	}

}
