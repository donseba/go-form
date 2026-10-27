// Package csrf holds the token stores of go-form's CSRF middleware: the
// Store, TokenConsumer and ExpiringStore interfaces and two in-memory
// stores. A store shared by several application instances implements Store,
// TokenConsumer to use each token once, atomically, and ExpiringStore when
// its tokens last another time than DefaultExpirationTime.
package csrf
