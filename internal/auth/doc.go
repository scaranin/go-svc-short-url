/*
Package auth handles user authentication using JSON Web Tokens (JWT).

It provides functionality for creating new user sessions, validating existing
tokens, and managing them through HTTP cookies. It defines the custom JWT claims
and the configuration necessary for signing and verifying tokens. The core logic
populates a UserID based on a request's cookie, making it available to the
application's handlers.
*/
package auth
