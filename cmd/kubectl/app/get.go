package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/client"
)

var (
	flagsGet flags

	cmdGet = &cobra.Command{
		Use:     "get <resources> | (<resource> <resource-name>)",
		Example: "get pods pod-nginx-1294a0bdf61d2ca\nget pods\n",
		Short:   "get resources by resource name",
		Args:    cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var url string
			if len(args) == 1 {
				resources := args[0]
				if !commandLineResources.Contains(resources) {
					fmt.Println("Unknown resource " + resources)
					return
				}
				resource := plural2singular[resources]
				url = baseUrl + fmt.Sprintf("/registry/%s/default", resource)

			} else if len(args) == 2 {
				resource := args[0]
				resourceName := args[1]
				if !commandLineResource.Contains(resource) {
					fmt.Println("Unknown resource " + resource)
					return
				}
				url = baseUrl + fmt.Sprintf("/registry/%s/default/%s", resource, resourceName)
			}
			listRes, err := client.Get(url)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			for _, res := range listRes {
				fmt.Println(res.ToString())
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(cmdGet)
}
