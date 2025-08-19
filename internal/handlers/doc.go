/*
Package handlers contains the HTTP request handlers that implement the application's
business logic.

This layer orchestrates interactions between incoming HTTP requests and the backend
services, such as the `storage` and `auth` packages. It is responsible for:
- Creating, retrieving, and deleting URLs.
- Handling single and batch operations.
- Processing different content types (text/plain and application/json).
- Performing health checks.
*/

package handlers
