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
