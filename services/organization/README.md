# `services/organization/`

This service manages the `Organization` model (also see `models/organization.go`).

Beware that `Organization` and `User` have a tight relationship: every `User` automatically has a "personal" `Organization` (so also check out `services/user/` if you need to edit something).
