package run

import (
	"fmt"
	"github.com/spf13/cobra"
)

var CmdRun = &cobra.Command{
	Use:   "run",
	Short: "Run project",
	Long:  "Run project. Example: bohe run",
	Run:   Run,
}

func Run(cmd *cobra.Command, args []string) {
	fmt.Println("hello world")
}
