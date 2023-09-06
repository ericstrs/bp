// Package storage provides a set of interfaces and implementations for
// data storage and retrieval. This package serves as a data abstraction
// layer for different storage backends likel YAML files, JSON or files,
// or databases.
//
// This package defines a storage interface that all storage
// implementations should follow. This allows core application to be
// withstand changing storage implementations.
package storage

// Storable is an interface for any data type that can be stored.
type Storable interface{}

// Storage is an interface for any storage mechanism (YAML, Database, etc.)
type Storage interface {
	Save(name string, data Storable) error
	Load(name string, into Storable) error
}
