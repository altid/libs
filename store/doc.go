/*
Package store implements a data store for Altid services.

Ramstore

Ramstore presents an in-memory only data store. This does not persist across restarts.

Logstore

Logstore presents an in-memory data store. Files ending with "/main" will be written to and read from real files which persist across restarts.

Special behaviour:

ctrl represents a special file, which when read from returns all the commands available for a service
when written to, it sends a command to the service
*/

package store
