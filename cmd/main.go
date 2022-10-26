package main

import (
	"mailfetcher/configs"
	"mailfetcher/internal/fetch"
)

func main() {
	fetch.Fetch(configs.Creds)

}
