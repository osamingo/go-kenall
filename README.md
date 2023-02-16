# go-kenall

[![CI](https://github.com/osamingo/go-kenall/workflows/CI/badge.svg)](https://github.com/osamingo/go-kenall/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/osamingo/go-kenall/branch/main/graph/badge.svg?token=gUDT8ydUMm)](https://codecov.io/gh/osamingo/go-kenall)
[![Go Report Card](https://goreportcard.com/badge/github.com/osamingo/go-kenall/v2)](https://goreportcard.com/report/github.com/osamingo/go-kenall/v2)
[![Go Reference](https://pkg.go.dev/badge/github.com/osamingo/go-kenall.svg)](https://pkg.go.dev/github.com/osamingo/go-kenall/v2)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/osamingo/go-kenall/blob/main/LICENSE)

## About

Unofficially [kenall](https://kenall.jp/) (ケンオール) client written by Go.

## Install

```shell
$ go get github.com/osamingo/go-kenall/v2@latest
```

## APIs supported by this library

- [郵便番号検索API](https://kenall.jp/docs/api-introduction/#%E9%83%B5%E4%BE%BF%E7%95%AA%E5%8F%B7%E6%A4%9C%E7%B4%A2api)
- [住所正規化API](https://kenall.jp/docs/API/postalcode/#%E4%BD%8F%E6%89%80%E6%AD%A3%E8%A6%8F%E5%8C%96%E6%A9%9F%E8%83%BD)
- [市区町村API](https://kenall.jp/docs/api-introduction/#%E5%B8%82%E5%8C%BA%E7%94%BA%E6%9D%91api)
- [日本の祝日API](https://kenall.jp/docs/API/holidays/)
- [法定休日確認API](https://kenall.jp/docs/API/businessday/)
- [自己IPアドレス確認API](https://kenall.jp/docs/API/whoami/#get-whoami)
- [法人番号検索API](https://kenall.jp/docs/api-introduction/#%E6%B3%95%E4%BA%BA%E7%95%AA%E5%8F%B7%E6%A4%9C%E7%B4%A2api)

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/osamingo/go-kenall/v2"
)

func main() {
	cli, err := kenall.NewClient(os.Getenv("KENALL_AUTHORIZATION_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	resAddr, err := cli.GetAddress(context.Background(), "1000001")
	if err != nil {
		log.Fatal(err)
	}
	addr := resAddr.Addresses[0]
	fmt.Println(addr.Prefecture, addr.City, addr.Town)
	// Output: 東京都 千代田区 千代田

	resCity, err := cli.GetCity(context.Background(), "13")
	if err != nil {
		log.Fatal(err)
	}
	city := resCity.Cities[0]
	fmt.Println(city.Prefecture, city.City)
	// Output: 東京都 千代田区

	res, err := cli.GetCorporation(context.Background(), "7000012050002")
	if err != nil {
		log.Fatal(err)
	}
	corp := res.Corporation
	fmt.Println(corp.PrefectureName, corp.CityName) 
	// Output: 東京都 千代田区
}
```

## Articles

- [ケンオール通信第1号](https://blog.kenall.jp/entry/kenall-newsletter-vol1)
  - This library has been featured on the official blog 🎉

## License

Released under the [MIT License](https://github.com/osamingo/go-kenall/blob/main/LICENSE).
