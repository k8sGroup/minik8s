package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/pkg/kubectl"
)

var (
	flagsGet flags

	cmdGet = &cobra.Command{
		Use:   "get <resource>",
		Short: "get resources by resource name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resource := args[0]
			var url string
			if flagsGet.name == "" {
				url = baseUrl + fmt.Sprintf("/registry/%s/%s", resource, flagsGet.namespace)
			} else {
				url = baseUrl + fmt.Sprintf("/registry/%s/%s/%s", resource, flagsGet.namespace, flagsGet.name)
			}
			bytes, err := kubectl.Get(url)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(string(bytes))
		},
	}
)

func init() {
	rootCmd.AddCommand(cmdGet)
	cmdGet.Flags().StringVarP(&flagsGet.namespace, "namespace", "n", "default", "namespace for a specific resource")
	cmdGet.Flags().StringVar(&flagsGet.name, "resource-name", "", "resource name")
}
