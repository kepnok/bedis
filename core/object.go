package core

import "strconv"

type Obj struct {
	TypeEncoding uint8
	Value        interface{}

	// Keeps track of the time when the object was last accessed at. Redis uses 24 bits to store this value but go doesnt support bitfield so here we are. We can optimize this later by having the first 8 bits contain the type and encoding. But for now it is what it is.
	LastAccessAt uint32
}

var (
	OBJ_TYPE_STRING     = uint8(0 << 4)
	OBJ_ENCODING_RAW    = uint8(0)
	OBJ_ENCODING_INT    = uint8(1)
	OBJ_ENCODING_EMBSTR = uint8(8)
)

func getTypeEncoding(value string) (uint8, uint8) {
	oType := OBJ_TYPE_STRING
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return oType, OBJ_ENCODING_INT
	}

	if len(value) <= 44 {
		return oType, OBJ_ENCODING_EMBSTR
	}

	return oType, OBJ_ENCODING_RAW
}
