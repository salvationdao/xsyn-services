# passport-server

[![Staging Deployment](https://github.com/ninja-syndicate/passport-server/actions/workflows/deploy-staging.yml/badge.svg)](https://github.com/ninja-syndicate/passport-server/actions/workflows/deploy-staging.yml)

[CD Docs](.github/workflows/README.md)

### Setup to use private repo

```bash
git config --global --add url."git@github.com:".insteadOf "https://github.com/"
export GOPRIVATE="github.com/ninja-software/*,github.com/ninja-syndicate/*"
```

### spinup

Windows spinup may have issues

```shell
make init / make init-docker
make serve / make serve-arelo
```

### envars

Majority of these don't need to be set for dev, if you want google/facebook/metamask auth then the relative ones will
need to be set.

```shell
PASSPORT_MORALIS_KEY
PASSPORT_ENVIRONMENT

PASSPORT_DATABASE_USER
PASSPORT_DATABASE_PASS
PASSPORT_DATABASE_HOST
PASSPORT_DATABASE_PORT
PASSPORT_DATABASE_NAME
PASSPORT_DATABASE_APPLICATION_NAME

PASSPORT_SENTRY_DSN_BACKEND
PASSPORT_SENTRY_SERVER_NAME
PASSPORT_SENTRY_SAMPLE_RATE

PASSPORT_HOST_URL_PUBLIC_FRONTEND
PASSPORT_API_ADDR

PASSPORT_COOKIE_SECURE

PASSPORT_MAIL_DOMAIN
PASSPORT_MAIL_APIKEY
PASSPORT_MAIL_SENDER

PASSPORT_JWT_ENCRYPT
PASSPORT_JWT_KEY
PASSPORT_JWT_EXPIRY_DAYS

PASSPORT_GOOGLE_CLIENT_ID
PASSPORT_METAMASK_SIGN_MESSAGE
PASSPORT_TWITCH_CLIENT_ID
PASSPORT_TWITCH_CLIENT_SECRET

PASSPORT_CORS_ALLOWED_ORIGINS
```
