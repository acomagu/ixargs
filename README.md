# ixargs: The xargs for interactive commands.

[![CircleCI](https://img.shields.io/circleci/project/github/acomagu/ixargs.svg?style=flat-square)](https://circleci.com/gh/acomagu/ixargs)

In case of xargs, user can't send anything to the child command when it prompts.

```
$ ls | xargs rm -i
rm: remove regular file 'a'? rm: remove regular file 'b'? rm: remove regular file 'c'?

$  # (´・ω・｀)
```

ixargs solves it.

```
$ ls | ixargs rm -i
rm: remove regular empty file 'a'? y
rm: remove regular empty file 'b'? y
rm: remove regular empty file 'c'? n

$  # (^_^)v
```

All options for xargs can be used.

## Installation

```
$ go get github.com/acomagu/ixargs
```

## How it works

Basic concept is to run:

```
xargs -I{} sh -c '</dev/tty $cmd {}'
```
