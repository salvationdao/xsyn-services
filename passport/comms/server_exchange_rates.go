package comms

import (
	"xsyn-services/passport/db"
	"xsyn-services/passport/payments"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

func (s *S) GetCurrentSupPrice(req GetCurrentSupPriceReq, resp *GetCurrentSupPriceResp) error {
	price := db.GetDecimal(db.KeySupsToUSD)
	if price.LessThanOrEqual(decimal.Zero) {
		exchangeRates, err := payments.FetchExchangeRates(true)
		if err != nil {
			return terror.Error(err, "Unable to fetch exchange rates.")
		}
		price = exchangeRates.SUPtoUSD
	}
	resp.PriceUSD = price

	return nil
}

func (s *S) GetCurrentRates(req GetExchangeRatesReq, resp *GetExchangeRatesResp) error {
	BNBtoUSD := db.GetDecimal(db.KeyBNBToUSD)
	ETHtoUSD := db.GetDecimal(db.KeyEthToUSD)
	SUPtoUSD := db.GetDecimal(db.KeySupsToUSD)

	if BNBtoUSD.LessThanOrEqual(decimal.Zero) || ETHtoUSD.LessThanOrEqual(decimal.Zero) || SUPtoUSD.LessThanOrEqual(decimal.Zero) {
		exchangeRates, err := payments.FetchExchangeRates(true)
		if err != nil {
			return terror.Error(err, "Unable to fetch exchange rates.")
		}
		BNBtoUSD = exchangeRates.BNBtoUSD
		ETHtoUSD = exchangeRates.ETHtoUSD
		SUPtoUSD = exchangeRates.SUPtoUSD
	}

	resp.BNBtoUSD = BNBtoUSD
	resp.ETHtoUSD = ETHtoUSD
	resp.SUPtoUSD = SUPtoUSD
	return nil
}
