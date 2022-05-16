package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/client"
)

type flags struct {
	namespace string
	name      string
}

var (
	flagsDel flags

	cmdDel = &cobra.Command{
		Use:   "del <resource> <resource-name>",
		Short: "delete resource",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			resource := args[0]
			resourceName := args[1]
			if !commandLineResource.Contains(resource) {
				fmt.Println("Unknown resource " + resource)
				return
			}
			url := baseUrl + fmt.Sprintf("/registry/%s/default/%s", resource, resourceName)
			err := client.Del(url)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println("Deleted")
		},
	}
)

func init() {
	rootCmd.AddCommand(cmdDel)
}
