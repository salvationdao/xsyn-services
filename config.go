package passport

type Config struct {
	CookieSecure        bool
	EncryptTokens       bool
	EncryptTokensKey    string
	TokenExpirationDays int
	AdminHostURL        string // The Admin Site URL used for CORS and links (eg: in the mailer)
	PublicHostURL       string // The Public Site URL used for CORS and links (eg: in the mailer)
	MobileHostURL       string // The Mobiel Site (flutter web) URL used for CORS
	MetaMaskSignMessage string // The message to see in the metamask signup flow, needs to match frontend
}
