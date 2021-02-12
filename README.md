# MessageRelayer

A message relayer that locally buffers messages received from a socket, preserves some of them and broadcasts to subscribers.

It handles the case where the socket is delayed for some time, or one of the subscribers cannot receive more messages.

Run tests:

```
make test
```

Run tests and coverage:

```
make cover
```