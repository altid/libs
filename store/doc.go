/*
Package store implements a data store for Altid services.

Ramstore

Ramstore presents an in-memory only data store. This does not persist across restarts.

Logstore

Logstore presents an in-memory data store. Files ending with "/main" will be written to and read from real files which persist across restarts.

*/

package store