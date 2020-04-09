/*Package tail mimicks the behaviour of `tail -f`

tail watches an events file for writes and sends out matching events, one for each line written.
It is careful to parse the data read in a separate goroutine, which helps assure all events are caught.

	ev, _ := tail.WatchEvents(ctx, path.Dir(dir), path.Base(dir))
	for event := range ev {
		fmt.Println(ev.Service, ev.Name)
		switch ev.EventType {
		case tail.FeedEvent:
			// [...]
		case tail.DocEvent:
			// [...]
		case tail.NotifyEvent:
			// [...]
		case tail.NoneEvent:
			// [...]
		}
	}

*/
package tail
