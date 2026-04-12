package mocks

import "github.com/stretchr/testify/mock"

// DB defines common database operations used by services.
type DB interface {
	Create(value any) error
	First(dest any, conds ...any) error
	Find(dest any, conds ...any) error
	Where(query any, args ...any) DB
}

// DBMock is a testify-backed mock implementation of DB.
type DBMock struct {
	mock.Mock
}

func (m *DBMock) Create(value any) error {
	args := m.Called(value)
	return args.Error(0)
}

func (m *DBMock) First(dest any, conds ...any) error {
	callArgs := []any{dest}
	callArgs = append(callArgs, conds...)
	args := m.Called(callArgs...)
	return args.Error(0)
}

func (m *DBMock) Find(dest any, conds ...any) error {
	callArgs := []any{dest}
	callArgs = append(callArgs, conds...)
	args := m.Called(callArgs...)
	return args.Error(0)
}

func (m *DBMock) Where(query any, conds ...any) DB {
	callArgs := []any{query}
	callArgs = append(callArgs, conds...)
	args := m.Called(callArgs...)
	if returned := args.Get(0); returned != nil {
		if db, ok := returned.(DB); ok {
			return db
		}
	}

	return m
}
