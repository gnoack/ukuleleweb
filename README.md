# Ukuleleweb

🌺 Ukuleleweb is a simple Wiki implementation in the style of the
original WikiWikiWeb / C2 wiki.

You can see it in action on [wiki.gnoack.org](https://wiki.gnoack.org/),
where you can also find [some documentation](https://wiki.gnoack.org/UkuleleWeb).

![Unit tests](https://github.com/gnoack/ukuleleweb/actions/workflows/go.yml/badge.svg)

## Few dependencies

The only dependencies are a markdown renderer and [Peter Bourgon's
`diskv` library](https://github.com/peterbourgon/diskv) for storage.

## Simple Storage

Wiki pages are simply stored in a directory, where file names are page
names and each file contains the page's markdown source. This way, you
can never lose your data, and it's also easy to put it under version
control.

## Features

* Recognizes WikiLinks in the classic WikiWikiWeb style (just CamelCasedWords).
* Recognizes go-link syntax (go/ukuleleweb)
* Displays reverse links at the bottom of each page. (Reverse links
  are recalculated when saving a page.)
* Each wiki page is a file on disk, it's easy to add small analysis
  scripts externally, or to generate new Wiki pages.

## Non-features

Ukuleleweb does not have user management, nor an authentication
mechanism. To keep your wiki confidential, place it behind a reverse
proxy.

## Run it

To run Ukuleleweb, create an empty directory where it can store data,
and point it to that directory using the `-store_dir` flag:

```
go run cmd/ukuleleweb/main.go -store_dir=/some/empty/directory
```

or, to install it:

```
$ go install github.com/gnoack/ukuleleweb/cmd/ukuleleweb@latest
$ go/bin/ukuleleweb -store_dir=/some/emtpy/directory
```
