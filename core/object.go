package core

import "strconv"

type Obj struct {
	TypeEncoding uint8
	Value        interface{}
	ExpiresAt    int64
}

var (
	OBJ_TYPE_STRING     = uint8(0 << 4)
	OBJ_ENCODING_RAW    = uint8(0)
	OBJ_ENCODING_INT    = uint8(1)
	OBJ_ENCODING_EMBSTR = uint8(8)
)

func getTypeEncoding(value string) (uint8, uint8) {
	oType := OBJ_TYPE_STRING
	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		return oType, OBJ_ENCODING_INT
	}

	if len(value) <= 44 {
		return oType, OBJ_ENCODING_EMBSTR
	}

	return oType, OBJ_ENCODING_RAW
}
