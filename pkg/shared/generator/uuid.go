package generator

import "github.com/google/uuid"

func NewUUIDv4() string {
	return uuid.NewString()
}

func NewUUIDv7() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}
