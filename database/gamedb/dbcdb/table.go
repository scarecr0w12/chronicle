package dbcdb

import (
	"fmt"
	"reflect"

	"github.com/Gophercraft/core/format/dbc"
)

type Table[T any] interface {
	Underlying() *dbc.Table
	Len() int
	Range(f func(cursor *T) bool) error
	Index(i int) (*T, error)
	ID(id int) (*T, error)
	StringRef(i int) (string, error)
}

func WrapTable[T any](t *dbc.Table) Table[T] {
	// Lookup by index is O(1), but lookup by ID is O(n) since we have to scan through the records to find the matching ID.
	// Optimize ID lookups by building a mapping of ID to index when we wrap the table.
	lookup := make(map[int]int)
	idx := 0
	err := t.Range(func(cursor *T) bool {
		id := reflect.ValueOf(cursor).Elem().Field(0).Int()
		lookup[int(id)] = idx
		idx++
		return true
	})
	if err != nil {
		panic("failed to wrap table: " + err.Error())
	}

	return &WrappedTable[T]{
		wrapped:    t,
		idxMapping: lookup,
	}
}

type WrappedTable[T any] struct {
	wrapped    *dbc.Table
	idxMapping map[int]int
}

func (w WrappedTable[T]) Len() int {
	return w.wrapped.Len()
}

func (w WrappedTable[T]) Underlying() *dbc.Table {
	return w.wrapped
}

func (w *WrappedTable[T]) Range(f func(cursor *T) bool) error {
	return w.wrapped.Range(func(cursor *T) bool {
		return f(cursor)
	})
}

func (w *WrappedTable[T]) ID(i int) (*T, error) {
	if idx, ok := w.idxMapping[i]; ok {
		return w.Index(idx)
	}

	return nil, fmt.Errorf("record not found: %d", i)

	//x := new(T)
	//if err := w.wrapped.ID(i, x); err != nil {
	//	return nil, err
	//}
	//return x, nil
}

func (w *WrappedTable[T]) Index(i int) (*T, error) {
	x := new(T)
	if err := w.wrapped.Index(i, x); err != nil {
		return nil, err
	}
	return x, nil
}

func (w *WrappedTable[T]) StringRef(i int) (string, error) {
	return w.wrapped.StringRef(i)
}
