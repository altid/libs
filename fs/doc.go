/*
Package fs is a library which allows you to write well constructed Altid services

Overview

fs aims to present a way to write canonical services, which behave correctly in all instances.
A service in Altid has the following structure on disk:

	/
		ctl
		events
		errors
		buffer1/
		buffer2/
		[...]


All buffer management occurs through writes to the ctl file. To open a buffer, for example, you write

	open mybuffer > /path/to/service/ctl

This in turn calls the Open function of a Handler, as described below.

Any errors encountered in the lifetime of a service will be logged to `errors`, and any events on any buffers will be logged to `events`
*/
package fs
