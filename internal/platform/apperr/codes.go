package apperr

// Stable public error codes. Clients branch on these values; they are part of
// the API contract and must not change once published.
const (
	CodeMalformedJSON       = "MALFORMED_JSON"
	CodeValidation          = "VALIDATION_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeAccountNotFound     = "ACCOUNT_NOT_FOUND"
	CodeAlreadyExists       = "ACCOUNT_ALREADY_EXISTS"
	CodeDuplicateExternalID = "DUPLICATE_EXTERNAL_ID"
	CodeInsufficientFunds   = "INSUFFICIENT_FUNDS"
	CodeSameAccountTransfer = "SAME_ACCOUNT_TRANSFER"
	CodeAccountBlocked      = "ACCOUNT_BLOCKED"
	CodeCurrencyMismatch    = "CURRENCY_MISMATCH"
	CodeIdempotency         = "ERRO_IDEMPOTENCIA"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeInternal            = "INTERNAL"
)
