# `services/user/`

This service manages the `User` model (also see `models/user.go`).

Beware that `User` and `Organization` have a tight relationship: every `User` automatically has a "personal" `Organization` (so also check out `services/organization/` if you need to edit something).
