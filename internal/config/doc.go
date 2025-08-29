/*
Package config manages the application's configuration.

It centralizes loading settings from multiple sources with a clear priority:
1. Environment Variables
2. Command-line Flags
3. Hard-coded Defaults

This package also includes a factory function, `CreateStore`, which selects and
initializes the appropriate storage backend (database, file, or in-memory) based
on the final configuration.
*/
package config
