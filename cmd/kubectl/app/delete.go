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
		Use:   "del <resource>",
		Short: "delete resource with flags",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resource := args[0]
			url := baseUrl + fmt.Sprintf("/registry/%s/%s/%s", resource, flagsDel.namespace, flagsDel.name)
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
	cmdDel.Flags().StringVarP(&flagsDel.namespace, "namespace", "n", "default", "namespace for a specific resource")
	cmdDel.Flags().StringVar(&flagsDel.name, "resource-name", "", "resource name")
	_ = cmdDel.MarkFlagRequired("resource-name")
}
