package until

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Cmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	outPut := &bytes.Buffer{}
	cmd.Stdout = outPut
	cmd.Stderr = outPut
	err := cmd.Run()
	if err != nil {
		fmt.Printf("execute command failed ,err %v\n", err)
		fmt.Println(string(outPut.Bytes()))
		return "", err
	}
	fmt.Println("command output", string(outPut.Bytes()))
	return string(outPut.Bytes()), nil
}
