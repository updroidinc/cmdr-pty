# cmdr-pty

A websocket wrapper for a pty written in Go.

Connect to a terminal emulator front-end via websocket or server-side tcp socket. Originally developed for use with [Terminal] in [UpDroid Commander].

## Usage

See `cmdr-pty --help` for usage:

```
usage: cmdr-pty [<flags>]

Flags:
  --help              Show help (also see --help-long and --help-man).
  -p, --protocol="websocket"  
                      specify websocket or tcp
  -a, --addr=":0"     IP:PORT or :PORT address to listen on
  -s, --size="25x80"  initial size for the tty

```

Use websocket mode for direct connection between client-side in browser and cmdr-pty. Use tcp mode if you have another server-side application to connect via tcp port.

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