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
	cmdDel = &cobra.Command{
		Use:   "del <resource> <resource-name>",
		Short: "delete resource",
		Args:  cobra.ExactArgs(2),
		Run:   deleteResource,
	}
)

func init() {
	rootCmd.AddCommand(cmdDel)
}

func deleteResource(cmd *cobra.Command, args []string) {
	resource := args[0]
	resourceName := args[1]
	if !commandLineResource.Contains(resource) {
		fmt.Println("Unknown resource " + resource)
		return
	}
	switch resource {
	case "replicaset":
		url := baseUrl + fmt.Sprintf("/registry/rsConfig/default/%s", resourceName)
		err := client.Del(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		break
	case "pod":
		url := baseUrl + fmt.Sprintf("/registry/podConfig/default/%s", resourceName)
		err := client.Del(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		break
	case "service":
		url := baseUrl + fmt.Sprintf("/registry/serviceConfig/default/%s", resourceName)
		err := client.Del(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		break
	case "dns":
		url := baseUrl + fmt.Sprintf("/registry/dnsAndTrans/default/%s", resourceName)
		err := client.Del(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		break
	case "deployment":
		url := baseUrl + fmt.Sprintf("/registry/%s/default/%s", resource, resourceName)
		err := client.Del(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		break
	case "autoscaler":
		url := baseUrl + fmt.Sprintf("/registry/%s/default/%s", resource, resourceName)
		err := client.Del(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		break
	}
	//if resource == "replicaset" {
	//	url := baseUrl + fmt.Sprintf("/registry/rsConfig/default/%s", resourceName)
	//	err := client.Del(url)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return
	//	}
	//} else {
	//	url := baseUrl + fmt.Sprintf("/registry/%s/default/%s", resource, resourceName)
	//	err := client.Del(url)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		return
	//	}
	//}
	fmt.Println("Deleted")
}
