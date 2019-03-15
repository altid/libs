# fslib - library to help write ubqt services

` go get -u github.com/ubqt-systems/fslib`

https://godoc.org/github.com/ubqt-systems/fslib

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
