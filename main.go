package main

import (
	"MemInspector/cmd"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

func main() {
	rogger.SetLevel(rogger.DEBUG)
	cmd.Execute()
}
