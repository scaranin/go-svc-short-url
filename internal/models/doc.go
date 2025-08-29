/*
Package models defines the core data structures and interfaces for the application.

It includes struct definitions for API requests and responses (Request, Response),
the main URL entity, and user-related data. A key component of this package is the
`Storage` interface, which defines the contract for all persistence layers, ensuring
that differen't storage backends (like database or file) can be used interchangeably.
*/
package models
