package config

import "time"

const (
	VERIFICATION_EMAIL_SUBJECT       = "OTP For Verification"
	VERIFICATION_DELETE_SUBJECT      = "OTP For Deletion"
	VERIFICATION_EMAIL_BODY          = "OTP: "
	VERIFICATION_OTP_EXPIRATION_TIME = 10 * time.Minute
)
