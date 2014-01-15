# envfmt

Format a null-separated file to create a sourceable file.

## install

```term
go get github.com/naaman/envfmt
```

## usage

Basic Output:

```term
$ envfmt -file /proc/2/environ -filter "TERM|COLUMNS|PWD|LINES|SHLVL|HOME|_|PATH|PS1"
export MULTI="A
B"
export DYNO="run.7532"
export PORT="31195"
```

Source Output:

```term
$ cat myenv
TERM=xterm-256colorMULTI=A
BCOLUMNS=178DYNO=run.7532PATH=/usr/local/bin:/usr/bin:/binPWD=/appPS1=\[\033[01;34m\]\w\[\033[00m\] \[\033[01;32m\]$ \[\033[00m\]LINES=51SHLVL=1HOME=/appPORT=31195_=/bin/bash
$ source <(envfmt -file myenv -filter "TERM|COLUMNS|PWD|LINES|SHLVL|HOME|_|PATH|PS1")
$ echo $MULTI
A
B
```
