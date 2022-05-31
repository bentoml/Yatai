package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kaniko/pkg/buildcontext"
)

func main() {
	buildCtx, err := buildcontext.GetBuildContext(os.Getenv("BUILD_CONTEXT"), buildcontext.BuildOptions{})
	if err != nil {
		panic(err)
	}
	srcCrx, err := buildCtx.UnpackTarFromBuildContext()
	if err != nil {
		panic(err)
	}
	fmt.Print(srcCrx)
}
