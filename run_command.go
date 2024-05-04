package main

import (
	"fmt"
	"os/exec"
)

func RunBittensorCliCommand(args []string) []byte {
	cmd := exec.Command("btcli", args...)

	stdout, err := cmd.Output()
	if err != nil {
		panic(err.Error())
	}

	return stdout
}

func RunPython3Command(args []string) []byte {
	fmt.Println(args)

	cmd := exec.Command("python3", args...)

	stdout, err := cmd.Output()
	if err != nil {
		panic(err.Error())
	}

	return stdout
}
