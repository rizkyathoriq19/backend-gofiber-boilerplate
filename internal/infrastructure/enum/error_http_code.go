package enum

type ErrorHTTPCodeType int

const (
	ERROR ErrorHTTPCodeType = iota
	SUCCESS
	INTERNAL_SERVER_ERROR

	ACTIVE
)

func (enum ErrorHTTPCodeType) Value() string {
	return [...]string{
		"error",
		"success",
		"Internal Server Error",
		"active",
	}[enum]
}

// v00-> credentials
// v01-> middleware