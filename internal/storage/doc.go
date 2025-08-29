/*
Package storage provides the concrete implementations of the `models.Storage` interface.

It contains different storage backends that can be used by the application:
  - `DBStorage`: A persistence layer using a PostgreSQL database. It handles database
    connections, schema creation, and all CRUD operations.
  - `FileStorageJSON`: A persistence layer that uses a local JSON file for storage,
    backed by an in-memory map for fast lookups.

The choice of which storage to use is determined by the application's configuration.
*/
package storage
