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
	path2 "path"
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
		fmt.Printf("Error analyzing file %s\n", path)
		return
	}
	kind := viper.GetString("kind")

	switch kind {
	case Deployment:
		if err := CaseDeployment(file, path, unmarshal); err != nil {
			return
		}
		break
	case Replicaset:
		if err := CaseReplicaset(file, path, unmarshal); err != nil {
			return
		}
		break
	case HorizontalPodAutoscaler:
		if err := CaseHPA(file, path, unmarshal); err != nil {
			return
		}
		break
	case GpuJob:
		if err := CaseGpuJob(file, path, unmarshal); err != nil {
			return
		}
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

func CaseDeployment(file []byte, path string, unmarshal func([]byte, any) error) error {
	deployment := object.Deployment{}
	err := unmarshal(file, &deployment)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s, %s\n", path, err.Error())
		return err
	}
	deployment.Complete()
	fmt.Printf("%+v\n", deployment)
	err = client.Put(baseUrl+"/registry/deployment/default/"+deployment.Metadata.Name, deployment)
	if err != nil {
		fmt.Printf("Error applying `file %s`.\n%s\n", path, err.Error())
		return err
	}
	return nil
}

func CaseReplicaset(file []byte, path string, unmarshal func([]byte, any) error) error {
	replicaset := object.ReplicaSet{}
	err := unmarshal(file, &replicaset)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s\n", path)
		return err
	}
	err = client.Put(baseUrl+path2.Join(config.RSConfigPrefix, replicaset.ObjectMeta.Name), replicaset)
	if err != nil {
		fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
		return err
	}
	return nil
}

func CaseHPA(file []byte, path string, unmarshal func([]byte, any) error) error {
	hpa := object.Autoscaler{}
	err := unmarshal(file, &hpa)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s\n", path)
		return err
	}
	err = client.Put(baseUrl+"/registry/autoscaler/default/"+hpa.Metadata.Name, hpa)
	if err != nil {
		fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
		return err
	}
	return nil
}

func CaseGpuJob(file []byte, path string, unmarshal func([]byte, any) error) error {
	uid := uuid.New().String()
	gpuJob := object.GPUJob{}
	err := unmarshal(file, &gpuJob)
	if err != nil {
		fmt.Printf("Error unmarshaling file %s\n", path)
		return err
	}
	gpuJob.Metadata.UID = uid
	zip, err := os.ReadFile(gpuJob.Spec.ZipPath)
	if err != nil {
		fmt.Printf("Error reading zip file `%s`\n.%s\n", gpuJob.Spec.ZipPath, err.Error())
		return err
	}
	jobKey := "job-" + uid
	jobZip := object.JobZipFile{
		Key:   jobKey,
		Slurm: gpuJob.GenerateSlurmScript(),
		Zip:   zip,
	}
	fmt.Println(string(jobZip.Slurm))
	err = client.Put(baseUrl+config.SharedDataPrefix+"/"+jobKey, jobZip)
	if err != nil {
		fmt.Printf("Error uploading file `%s`\n.%s\n", gpuJob.Spec.ZipPath, err.Error())
		return err
	}
	err = client.Put(baseUrl+"/registry/job/default/"+jobKey, gpuJob)
	if err != nil {
		fmt.Printf("Error applying file `file%s`\n.%s\n", path, err.Error())
		return err
	}
	return nil
}
