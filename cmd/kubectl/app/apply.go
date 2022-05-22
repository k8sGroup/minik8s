package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"os"
	"strings"
)

const (
	Deployment              string = "Deployment"
	Replicaset              string = "Replicaset"
	HorizontalPodAutoscaler string = "HorizontalPodAutoscaler"
	Test                    string = "Test"
	GpuJob                  string = "GpuJob"
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
		CaseDeployment(file, path, unmarshal)
		break
	case Replicaset:
		CaseReplicaset(file, path, unmarshal)
		break
	case HorizontalPodAutoscaler:
		CaseHPA(file, path, unmarshal)
		break
	case GpuJob:
		CaseGpuJob(file, path, unmarshal)
		break
	case Test:
		err = client.Put(baseUrl+"/registry/test/default/test1", "{test:\"test\"}")
		if err != nil {
			fmt.Println(err.Error())
		}
		break
	case "":
		fmt.Printf("kind field is unspecified\n")
		return
	default:
		fmt.Printf("Unsupported kind %s\n", kind)
		return
	}
	fmt.Println("Applied!")
}

func CaseDeployment(file []byte, path string, unmarshal func([]byte, any) error) {
	deployment := object.Deployment{}
	err := unmarshal(file, &deployment)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s, %s\n", path, err.Error())
		return
	}
	deployment.Complete()
	fmt.Printf("%+v\n", deployment)
	err = client.Put(baseUrl+"/registry/deployment/default/"+deployment.Metadata.Name, deployment)
	if err != nil {
		fmt.Printf("Error applying `file %s`.\n%s\n", path, err.Error())
		return
	}
}

func CaseReplicaset(file []byte, path string, unmarshal func([]byte, any) error) {
	replicaset := object.ReplicaSet{}
	err := unmarshal(file, &replicaset)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s\n", path)
		return
	}
	err = client.Put(baseUrl+"/registry/rs/default/"+replicaset.ObjectMeta.Name, replicaset)
	if err != nil {
		fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
		return
	}
}

func CaseHPA(file []byte, path string, unmarshal func([]byte, any) error) {
	hpa := object.Autoscaler{}
	err := unmarshal(file, &hpa)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s\n", path)
		return
	}
	err = client.Put(baseUrl+"/registry/autoscaler/default/"+hpa.Metadata.Name, hpa)
	if err != nil {
		fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
		return
	}
}

func CaseGpuJob(file []byte, path string, unmarshal func([]byte, any) error) {
	uid := uuid.New().String()
	gpuJob := object.GPUJob{}
	err := unmarshal(file, &gpuJob)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s\n", path)
		return
	}
	gpuJob.Metadata.UID = uid
	zip, err := os.ReadFile(gpuJob.Spec.ZipPath)
	if err != nil {
		fmt.Printf("Error reading zip file `%s`\n.%s\n", gpuJob.Spec.ZipPath, err.Error())
		return
	}
	jobKey := "job-" + uid
	jobZip := object.JobZipFile{
		Key:   jobKey,
		Slurm: gpuJob.GenerateSlurmScript(),
		Zip:   zip,
	}
	err = client.Put(baseUrl+config.SharedDataPrefix+"/"+jobKey, jobZip)
	if err != nil {
		fmt.Printf("Error uploading file `%s`\n.%s\n", gpuJob.Spec.ZipPath, err.Error())
		return
	}
	err = client.Put(baseUrl+"/registry/job/default/"+jobKey, gpuJob)
	if err != nil {
		fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
		return
	}
}
