package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"minik8s/object"
	"minik8s/pkg/client"
	"os"
	"strings"
)

const (
	Deployment              string = "Deployment"
	Replicaset              string = "Replicaset"
	HorizontalPodAutoscaler string = "HorizontalPodAutoscaler"
)

var (
	cmdApply = &cobra.Command{
		Use:   "apply <pathname>",
		Short: "apply to minik8s with a yaml file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			fmt.Println("Path : ", path)
			analyzeFile(path)
		},
	}
)

func init() {
	rootCmd.AddCommand(cmdApply)
	cmdDel.Flags().StringVarP(&flagsDel.namespace, "namespace", "n", "default", "namespace for a specific resource")
	cmdDel.Flags().StringVar(&flagsDel.name, "resource-name", "", "resource name")
	_ = cmdDel.MarkFlagRequired("resource-name")
}

func analyzeFile(path string) {
	var unmarshal func([]byte, any) error
	if strings.HasSuffix(path, "json") {
		viper.SetConfigType("json")
		unmarshal = json.Unmarshal
	} else if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {
		viper.SetConfigType("yaml")
		unmarshal = yaml.Unmarshal
	} else {
		fmt.Printf("Unsupported type! Apply a yaml or json file!\n")
		return
	}

	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s\n", path)
		return
	}
	err = viper.ReadConfig(bytes.NewReader(file))
	if err != nil {
		fmt.Printf("Error reading file %s\n", path)
		return
	}
	kind := viper.GetString("Kind")

	switch kind {
	case Deployment:
		deployment := object.Deployment{}
		err = unmarshal(file, &deployment)
		if err != nil {
			fmt.Printf("Error unmarshaling file %s\n", path)
			return
		}
		deployment.Complete()
		err = client.Put(baseUrl+"/registry/deployment/default/"+deployment.Metadata.Name, deployment)
		if err != nil {
			fmt.Printf("Error applying `file %s`.\n%s\n", path, err.Error())
			return
		}
		break
	case Replicaset:
		replicaset := object.ReplicaSet{}
		err = unmarshal(file, &replicaset)
		if err != nil {
			fmt.Printf("Error unmarshaling file %s\n", path)
			return
		}
		err = client.Put(baseUrl+"/registry/rs/default/"+replicaset.ObjectMeta.Name, replicaset)
		if err != nil {
			fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
			return
		}
		break
	case HorizontalPodAutoscaler:
		hpa := object.Autoscaler{}
		err = unmarshal(file, &hpa)
		if err != nil {
			fmt.Printf("Error unmarshaling file %s\n", path)
			return
		}
		err = client.Put(baseUrl+"/registry/rs/default/"+hpa.Metadata.Name, hpa)
		if err != nil {
			fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
			return
		}
		break
	case "":
		fmt.Printf("kind is nspecified\n")
		return
	default:
		fmt.Printf("Unknown kind %s\n", kind)
		return
	}
	fmt.Println("Applied!")
}
