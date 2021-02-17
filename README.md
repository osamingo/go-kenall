# go-kenall

[![CI](https://github.com/osamingo/go-kenall/workflows/CI/badge.svg)](https://github.com/osamingo/go-kenall/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/osamingo/go-kenall/branch/main/graph/badge.svg?token=gUDT8ydUMm)](https://codecov.io/gh/osamingo/go-kenall)
[![Go Report Card](https://goreportcard.com/badge/github.com/osamingo/go-kenall)](https://goreportcard.com/report/github.com/osamingo/go-kenall)
[![Go Reference](https://pkg.go.dev/badge/github.com/osamingo/go-kenall.svg)](https://pkg.go.dev/github.com/osamingo/go-kenall)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/osamingo/go-kenall/blob/main/LICENSE)

## About

Unofficially [kenall](https://kenall.jp/) (ケンオール) client written by Go.

## Install

```shell
$ go get -u github.com/osamingo/go-kenall
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/osamingo/go-kenall"
)

func main() {

	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := cli.GetAddress(context.Background(), "1000001")
	if err != nil {
		log.Fatal(err)
	}

	addr := res.Addresses[0]
	fmt.Println(addr.Prefecture, addr.City, addr.Town)
	// Output: 東京都 千代田区 千代田
}
```

## Articles

- [ケンオール通信第1号](https://blog.kenall.jp/entry/kenall-newsletter-vol1)
  - This library has been featured on the official blog 🎉

## License

Released under the [MIT License](https://github.com/osamingo/go-kenall/blob/main/LICENSE).
