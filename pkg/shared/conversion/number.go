package conversion

import "fmt"

/*
IntToUint32 converts an int to uint32, returning an error if the value is
out of range.
*/
func IntToUint32(i int) (uint32, error) {
	if i < 0 || i >= 1<<32 {
		return 0, fmt.Errorf("int value %d is out of range for uint32", i)
	}
	return uint32(i), nil
}

/*
IntToUint16 converts an int to uint16, returning an error if the value is
out of range.
*/
func IntToUint16(i int) (uint16, error) {
	if i < 0 || i >= 1<<16 {
		return 0, fmt.Errorf("int value %d is out of range for uint16", i)
	}
	return uint16(i), nil
}

/*
IntToUint8 converts an int to uint8, returning an error if the value is
out of range.
*/
func IntToUint8(i int) (uint8, error) {
	if i < 0 || i >= 1<<8 {
		return 0, fmt.Errorf("int value %d is out of range for uint8", i)
	}
	return uint8(i), nil
}
