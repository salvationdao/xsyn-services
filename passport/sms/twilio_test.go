package sms_test

import (
	"os"
	"testing"
	"xsyn-services/passport/sms"
)

func TestTwilio_SendSMS(t *testing.T) {
	accountSid := os.Getenv("PASSPORT_TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("PASSPORT_TWILIO_API_KEY")
	apiSecret := os.Getenv("PASSPORT_TWILIO_API_SECRET")
	fromNumber := os.Getenv("PASSPORT_SMS_FROM_NUMBER")
	toNumber := "+61478147822"

	twil, err := sms.NewTwilio(
		accountSid,
		apiKey,
		apiSecret,
		fromNumber,
		"staging",
	)
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to created new twilio")
	}

	_, err = twil.Lookup(toNumber)
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to lookup number %s", toNumber)
	}

	err = twil.SendSMS(toNumber, "Test message from xsyn passport")
	if err != nil {
		t.Log(err)
		t.Fatalf("failed to send number %s", toNumber)
	}
}
