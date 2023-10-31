package library

import (
	"context"
	"errors"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
	ErrUnavaliable       = errors.New("library service unavailable")
)

type Config struct {
	Host     string
	Port     string
	MaxFails uint32 `mapstructure:"max_fails"`
}

type Client interface {
	GetLibraries(context.Context, string, uint64, uint64) (Infos, error)
	GetLibrariesByIDs(context.Context, []string) (Infos, error)
	GetBooks(context.Context, string, bool, uint64, uint64) (Books, error)
	GetBooksByIDs(context.Context, []string) (Books, error)
	ObtainBook(context.Context, string, string) (ReservedBook, error)
	ReturnBook(context.Context, string, string) (Book, error)
}
