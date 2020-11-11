# `services/secret/`

This service manages secrets (in other apps also known as access tokens or API keys). Secrets either belong to a user (`models.UserSecret`) or a service (`models.ServiceSecret`). Secrets are used to authenticate requests (including normal frontend requests).

The package includes a cache for checking secrets that can be used to authenticate every request. It returns a `models.Secret` interface, which handily includes several owner-related fields joined from other tables (eg. the secret owner's quotas).
