.. -*- mode: rst -*-

===========
 goreplace
===========

goreplace is a simple utility which can be used as replacement for grep + sed
combination in one of most popular cases - find files, which contain something,
possibly replace this with something else.

Installation
------------

It's suited to be installed via ``go`` tool, so just do usual thing::

  go get github.com/piranha/goreplace

And you should be done. You'll have ``$GOPATH`` set to some directory for this
to work (it'll put sources and generated binary there).

I prefer name ``gr`` to ``goreplace``, so I link ``gr`` somewhere in my path
(usually in ``~/bin``) to ``$GOPATH/bin/goreplace``.

Usage
-----

Usage is pretty simple, you can just run ``gr`` to see help on
options. Basically you just supply regexp (or a simple string - it's a regexp
always as well) as an argument and goreplace will search for it in all files
starting from the current directory, just like this::

  gr somestring

Some directories and files can be ignored by default (``gr`` is looking for your
``.hgignore``/``.gitignore`` in parent directories), just run ``gr`` without any
arguments to see help message - it contains information about them.

If you need to replace found strings with something, just pass ``-r
replacement`` option and they will be replaced in-place. No backups are made
(not that you need them, right? You're using version control, aren't you?).
Unfortunately only plain strings are supported as replacement, no regexp
submatch support yet (planned, though).

