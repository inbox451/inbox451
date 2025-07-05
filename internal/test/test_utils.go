package test

import (
	"github.com/google/uuid"
)

func StaticTestUUID() string {
	return "e98b45a1-21f7-4ac7-8e49-3d20bfb8c10d"
}

func RandomTestUUID() string {
	return uuid.New().String()
}
