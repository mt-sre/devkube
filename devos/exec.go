package devos

import (
	"k8s.io/utils/exec"
)

type Exec = exec.Interface

func RealExecIfUnset(e Exec) Exec {
	if e == nil {
		return exec.New()
	}

	return e
}
