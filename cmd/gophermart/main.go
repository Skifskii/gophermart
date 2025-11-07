package main

import "gophermart/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
