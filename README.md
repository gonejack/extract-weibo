# extract-weibo
This command line parses .htmls saved by [saveurls](https://github.com/gonejack/saveurls) from `m.weibo.cn`.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gonejack/extract-weibo)
![Build](https://github.com/gonejack/extract-weibo/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/github/license/gonejack/extract-weibo.svg?color=blue)](LICENSE)

### Install
```shell
> go get github.com/gonejack/extract-weibo
```

### Usage
```shell
> extract-weibo *.html
```
```
Flags:
  -h, --help       Show context-sensitive help.
  -c, --convert    Convert weibo.com links to m.weibo.cn.
  -v, --verbose    Verbose printing.
```
