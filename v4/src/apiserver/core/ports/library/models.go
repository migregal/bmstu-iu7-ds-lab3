package library

import "github.com/migregal/bmstu-iu7-ds-lab2/pkg/collections"

type Info struct {
	ID      string
	Name    string
	Address string
	City    string
}

type Infos collections.Countable[Info]

type Book struct {
	ID        string
	Name      string
	Author    string
	Genre     string
	Condition string
	Available uint64
}

type Books collections.Countable[Book]

type ReservedBook struct {
	Book    Book
	Library Info
}
