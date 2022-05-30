package app

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"path"
)

type rsStatus struct {
	Actual int32 `json:"actual" yaml:"actual"`
	Expect int32 `json:"expect" yaml:"expect"`
}

type rsWithVersion struct {
	replicaset    object.ReplicaSet
	createVersion int64
}

type beautifiedReplicaset struct {
	name    string
	uid     string
	podName string
	status  rsStatus
}

type beautifiedDeployment struct {
	name string
	rs   *beautifiedReplicaset
}

func BRSHeader() string {
	return "REPLICASET-NAME\tUUID\tPOD-TEMPLATE-NAME\tEXPECT\tACTUAL\n"
}

func BDPHeader() string {
	return "DEPLOYMENT-NAME\tREPLICASET-NAME\tUUID\tPOD-TEMPLATE-NAME\tEXPECT\tACTUAL\n"
}

func (brs *beautifiedReplicaset) ToString() string {
	uid := brs.uid
	if uid == "" {
		uid = "null"
	}
	return fmt.Sprintf("%s\t%s\t%s\t%d\t%d\n", brs.name, uid, brs.podName, brs.status.Expect, brs.status.Actual)
}

func (bdp *beautifiedDeployment) ToString() string {
	uid := bdp.rs.uid
	if uid == "" {
		uid = "null"
	}
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%d\n", bdp.name, bdp.rs.name, uid, bdp.rs.podName, bdp.rs.status.Expect, bdp.rs.status.Actual)
}

var (
	cmdGet = &cobra.Command{
		Use:     "get <resources> | (<resource> <resource-name>)",
		Example: "get pods pod-nginx-1294a0bdf61d2ca\nget pods\n",
		Short:   "get resources by resource name",
		Args:    cobra.RangeArgs(1, 2),
		Run:     getHandler,
	}
)

func init() {
	rootCmd.AddCommand(cmdGet)
}

func getHandler(cmd *cobra.Command, args []string) {
	var name string
	if len(args) == 2 {
		name = args[1]
	}

	switch args[0] {
	case "svc":
		caseService(name)
		return
	case "replicaset":
		caseReplicaset(name)
		return
	case "deployment":
		caseDeployment(name)
		return
	default:
		fmt.Println("Unknown resource ", args[0])
	}
}

func caseService(name string) {
	//TODO
	fmt.Println(name)
	fmt.Println("Not supported yet")
}

func caseDeployment(name string) {

	url := baseUrl + config.RSConfigPrefix
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	dm2rs := make(map[string]*rsWithVersion)
	for _, res := range listRes {
		var rs object.ReplicaSet
		err = json.Unmarshal(res.ValueBytes, &rs)
		if err != nil {
			continue
		}
		rsTmp := &rsWithVersion{
			createVersion: res.CreateVersion,
			replicaset:    rs,
		}
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.Kind == "Deployment" && ownerRef.Name != "" {
				rsOld, ok := dm2rs[ownerRef.Name]
				if !ok {
					dm2rs[ownerRef.Name] = rsTmp
				} else if rsOld.createVersion < rsTmp.createVersion {
					dm2rs[ownerRef.Name] = rsTmp
				}
			}
		}
	}

	url = baseUrl + path.Join("/registry/deployment/default", name)
	listRes, err = client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var beautifiedDeployments []*beautifiedDeployment
	for _, res := range listRes {
		var deployment object.Deployment
		err = json.Unmarshal(res.ValueBytes, &deployment)
		if err != nil {
			continue
		}
		rs, ok := dm2rs[deployment.Metadata.Name]
		if ok {
			status := rsStatus{}
			data, err := client.GetWithParams(baseUrl+config.RS_POD, map[string]string{"rsName": rs.replicaset.Name, "uid": rs.replicaset.UID})
			if err != nil {
				continue
			}
			err = json.Unmarshal(data, &status)
			if err != nil {
				continue
			}
			beautifiedDeployments = append(beautifiedDeployments, &beautifiedDeployment{
				name: deployment.Metadata.Name,
				rs: &beautifiedReplicaset{
					name:    rs.replicaset.Name,
					uid:     rs.replicaset.UID,
					podName: rs.replicaset.Spec.Template.Name,
					status:  status,
				},
			})
		}
	}
	fmt.Print(BDPHeader())
	for _, bdp := range beautifiedDeployments {
		fmt.Print(bdp.ToString())
	}
}

func caseReplicaset(name string) {
	url := baseUrl + path.Join(config.RSConfigPrefix, name)
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var beautifiedReplicasets []*beautifiedReplicaset
	for _, res := range listRes {
		var rs object.ReplicaSet
		err = json.Unmarshal(res.ValueBytes, &rs)
		if err != nil {
			continue
		}
		brs := beautifiedReplicaset{
			name:    rs.Name,
			uid:     rs.UID,
			podName: rs.Spec.Template.Name,
		}
		status := rsStatus{}
		data, err := client.GetWithParams(baseUrl+config.RS_POD, map[string]string{"rsName": rs.Name, "uid": rs.UID})
		if err != nil {
			continue
		}
		err = json.Unmarshal(data, &status)
		if err != nil {
			continue
		}
		brs.status = status
		beautifiedReplicasets = append(beautifiedReplicasets, &brs)
	}
	fmt.Print(BRSHeader())
	for _, brs := range beautifiedReplicasets {
		fmt.Print(brs.ToString())
	}
}
