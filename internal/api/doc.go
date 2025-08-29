/*
Package api is responsible for initializing the application's HTTP router.

It uses the `chi` router to define all API endpoints and maps them to the
corresponding handler methods in the `handlers` package. It also applies
global middleware, such as logging and Gzip compression, to all incoming
requests.
*/
package api
