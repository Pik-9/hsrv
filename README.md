# hsrv
**hsrv** is a very simple http server written in Go that serves the files of a given directory.
It is meant to be used as a helper tool during web development on the localhost.

## Usage
```bash
hsrv
```
Calling `hsrv` without any parameters will cause it to serve the files from the _current working
directory_ on port 8080.

```bash
hsrv -p 3000
```
You can pass a port to listen to with the `-p` flag.

```bash
hsrv /path/to/directory
```
```bash
hsrv sub/directory
```
You can specify the path to the files to serve either as an absolute path or as a sub path
of the current working directory.

`hsrv` will write all http requests to stdout and can be shutdown by sending a **SIGTERM**,
**SIGQUIT** or a **SIGINT** ([Ctrl]-[C]).

## Building
```bash
go build
```
Simple as that.
