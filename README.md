# fslib

Fslib is a library designed to ease the creation of ubqt-style file servers in Golang.
(Link to godoc will eventually be linked here when the documentation is more complete)


## Features

fslib manages the lifetime and behavior of a control file-interface to a running service. Writing and reading to a given control file will trigger actions, and this library will manage how those actions manifest in other files related to the service - tabs shows files you've opened (example, echo open foobar >> /tmp/ubqt/myservice/ctrl), events related to changes to managed files are written to the events file, (example /tmp/ubqt/myservice/events) and creating a new buffer with a backing log.

If logs are not desired as a configuration-time option, the special value of `log=none` can be used.


## Installation

` go get -u github.com/ubqt-systems/fslib`


## Bugs

fslib attempts to create a canonical interface to a ubqt service. However, it is very new, and bugs are likely. Please open issues!

ctrl and input files currently only work in append-only mode. This is the intended behavior, as it provides command and input history; but it must be stated that when you do truncated writes the behavior is undefined.

```
# for example
echo 'open #mychannel' >> /tmp/ubqt/irc.freenode.net/ctrl
echo 'close #mychannel' >> /tmp/ubqt/irc.freenode.net/ctrl

# will work correctly, while 

echo 'open #mychannel' > /tmp/ubqt/irc.freenode.net/ctrl
echo 'close #mychannel' > /tmp/ubqt/irc.freenode.net/ctrl

# will not work correctly.
```

Any fixes for this are welcome, please set build flags for the supported GOOS; an inotify based solution should work well for systems which support it.
