package main

import "gitlab.com/jgillich/autominer/cmd"

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
