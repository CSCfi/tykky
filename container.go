package tykky

import (
	"fmt"
	"os"
	"strings"
)

const RUNTIME_UNCHECKED = 0
const RUNTIME_CHECKED = 1

var RUNTIME_CHECK_STATE = RUNTIME_UNCHECKED
var apptainerIsAvailable = false

func hasApptainer() bool {
	if RUNTIME_CHECK_STATE == RUNTIME_UNCHECKED {
		apptainerIsAvailable = CommandExists("apptainer")
		RUNTIME_CHECK_STATE = RUNTIME_CHECKED
	}
	return apptainerIsAvailable
}

func getContainerRuntime() string {
	if hasApptainer() {
		return "apptainer"
	} else {
		return "singularity"
	}
}

func GetApptainerContainer(containerSrc string, containerDest string, shareContainer bool) {
	if PathExists(containerSrc) {
		if shareContainer {
			os.Symlink(containerDest, containerSrc)
		} else {
			CopyFile(containerSrc, containerDest)
		}
	} else {
		fc := fmt.Sprintf("%s --silent pull %s %s", getContainerRuntime(), containerDest, containerSrc)
		RunCommand(fc, PlainOutputWriterNoLog)
	}
}

type ApptainerContainerInstance struct {
	ContainerPath string
	BindMounts    []string
}

func (cont ApptainerContainerInstance) RunInContainer(command string, scrollOutput bool) {
	fc := fmt.Sprintf("%s --silent exec -B %s %s %s", getContainerRuntime(), strings.Join(cont.BindMounts, ","), cont.ContainerPath, command)
	if scrollOutput {
		RunCommand(fc, ScrollingOutputWriterNoLog)
	} else {
		RunPlainOutput(fc)
	}
}
