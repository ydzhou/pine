# (P)ine (I)s (N)ot (E)ditor

Pine is a terminal editor. It is inspired by [uemacs](https://github.com/torvalds/uemacs) and [nano](https://www.nano-editor.org/)

## Features

It is a very basic editor and try to keep the same style as uemacs or nano in a retro way. Still it supports some modern features to smooth your experience.

* Multiple buffers support: you can open and edit multiple files at the same time.
* Native UTF-8 support: you can type and view Chinese, Japanese, Arabic and other language characters.

## Roadmap

* Auto indention
* Mouse support for scrolling and moving the cursor
* Copy and paste within the editor
* Undo
* Search text

## Install

`golang` is the only requirement.

run `make build`

move compiled binary file `pine` to any directory you prefer, e.g. `/usr/local/bin` and include in the `PATH`
