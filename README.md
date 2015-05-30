# cmdr-pty

A websocket wrapper for a pty written in Go.

Connect to a terminal emulator front-end via websocket. Originally developed for use with [Terminal] in [UpDroid Commander].

## Usage

See `cmdr-pty --help` for usage:

```
Usage of ./cmdr-pty:
  -addr=":0": IP:PORT or :PORT address to listen on
  -size="24x80": initial size for the tty
```

Resize the terminal by entering (rows)x(columns) into stdin.

If using with [Terminal] 0.1.0, make sure you `export TERM=vt100` before running `cmdr-pty` since it does not yet handle some vt102/xterm escape sequences.


## Contribute

Pull requests welcome! Though, I reserve the right to review and/or reject them at will.
Can also file issues with the issue tracker.

### TODO:

- Move resize and other command handling to another websocket endpoint for applications where access to the process' stdin is limited.
- Add more control options/enhancements.

## Acknowledgements

Heavily inspired by the [pty.js] project by (chjj) Christopher Jeffrey and other fork-pty projects.

[Terminal]: https://github.com/updroidinc/terminal/
[UpDroid Commander]: http://updroid.com/commander/
[pty.js]: https://github.com/chjj/pty.js/