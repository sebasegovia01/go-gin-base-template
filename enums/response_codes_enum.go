package enums

type ResultStatus string

const (
	OK      ResultStatus = "OK"
	ERROR   ResultStatus = "ERROR"
	WARNING ResultStatus = "WARNING"
)

type CanonicalErrorType string

const (
	NEG CanonicalErrorType = "NEG"
	TEC CanonicalErrorType = "TEC"
	SEG CanonicalErrorType = "SEG"
)
