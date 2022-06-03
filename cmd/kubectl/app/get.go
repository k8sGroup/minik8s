package app

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"path"
	"strings"
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
type beautifiedPod struct {
	Name     string
	Ctime    string
	PodIp    string
	NodeName string
	Status   string
}

type beautifiedAutoscaler struct {
	name          string
	refKind       string
	refName       string
	minReplicas   int32
	maxReplicas   int32
	scaleInterval int32
	metrics       []object.Metric
}

func BASHeader() string {
	return "Name\tRefKind\tRefName\tMin\tMax\tInterval\tMetrics\n"
}

func (b *beautifiedAutoscaler) ToString() string {
	metricsList := make([]string, len(b.metrics))
	for i, str := range b.metrics {
		metricsList[i] = str.ToString()
	}
	metrics := "{" + strings.Join(metricsList, ", ") + "}"
	result := fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%d\t%s\n", b.name, b.refKind, b.refName, b.minReplicas, b.maxReplicas, b.scaleInterval, metrics)
	return result
}

func (bPod *beautifiedPod) ToString() string {
	result := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", bPod.Name, bPod.Ctime, bPod.PodIp, bPod.NodeName, bPod.Status)
	return result
}

type beautifiedDeployment struct {
	name string
	rs   *beautifiedReplicaset
}
type beautifiedNode struct {
	Name          string
	Ctime         string
	MasterIp      string
	NodeIp        string
	NodeIpAndMask string
}

func (bNode *beautifiedNode) ToString() string {
	result := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", bNode.Name, bNode.Ctime, bNode.MasterIp, bNode.NodeIp, bNode.NodeIpAndMask)
	return result
}

type beautifiedService struct {
	Name      string
	Ctime     string
	ClusterIP string
	Ports     []PortAndProtocol
	Status    string
}
type PortAndProtocol struct {
	Port     string
	Protocol string
}

func (bSvc *beautifiedService) ToString() string {
	ports2string := "["
	for _, port := range bSvc.Ports {
		ports2string += port.Port + ":" + port.Protocol + " "
	}
	ports2string += "]"
	result := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", bSvc.Name, bSvc.Ctime, bSvc.ClusterIP, ports2string, bSvc.Status)
	return result
}

type beautifiedDnsAndTrans struct {
	Name      string
	Ctime     string
	Host      string
	Path2Svcs []Path2Svc
	Status    string
}

type Path2Svc struct {
	Path string
	Svc  string
}

func (bDnsAndTrans *beautifiedDnsAndTrans) ToString() string {
	path2Svcs := "["
	for _, path2Svc := range bDnsAndTrans.Path2Svcs {
		path2Svcs += path2Svc.Path + ":" + path2Svc.Svc + " "
	}
	path2Svcs += "]"
	result := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", bDnsAndTrans.Name, bDnsAndTrans.Ctime, bDnsAndTrans.Host, path2Svcs, bDnsAndTrans.Status)
	return result
}
func BRSHeader() string {
	return "REPLICASET-NAME\tUUID\tPOD-TEMPLATE-NAME\tEXPECT\tACTUAL\n"
}

func BDPHeader() string {
	return "DEPLOYMENT-NAME\tREPLICASET-NAME\tUUID\tPOD-TEMPLATE-NAME\tEXPECT\tACTUAL\n"
}

func NODEHeader() string {
	return "NodeName\tCtime\tMasterIp\tNodeIp\tNodeIpAndMask\n"
}
func PODHeader() string {
	return "PodName\tCtime\tPodIp\tNodeName\tStatus\n"
}
func SERVICEHeader() string {
	return "ServiceName\tCtime\tClusterIp\tPorts\tStatus\n"
}
func DnsAndTransHeader() string {
	return "DnsAndTransName\tCtime\tHost\tPath2Svcs\tStatus\n"
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
	case "service":
		caseService(name)
		return
	case "replicaset":
		caseReplicaset(name)
		return
	case "deployment":
		caseDeployment(name)
		return
	case "pod":
		casePod(name)
		return
	case "node":
		caseNode(name)
		return
	case "dnsAndTrans":
		caseDnsAndTrans(name)
		return
	case "autoscaler":
		caseAutoscaler(name)
		return
	case "job":
		caseJob(name)
		return
	default:
		fmt.Println("Unknown resource ", args[0])
	}
}
func caseDnsAndTrans(name string) {
	url := baseUrl + path.Join(config.DnsAndTransPrefix, name)
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var results []*beautifiedDnsAndTrans
	for _, res := range listRes {
		dnsAndTrans := &object.DnsAndTrans{}
		err = json.Unmarshal(res.ValueBytes, dnsAndTrans)
		if err != nil {
			continue
		}
		bDnsAndTrans := &beautifiedDnsAndTrans{
			Name:   dnsAndTrans.MetaData.Name,
			Ctime:  dnsAndTrans.MetaData.Ctime,
			Host:   dnsAndTrans.Spec.Host,
			Status: dnsAndTrans.Status.Phase,
		}
		var path2Svcs []Path2Svc
		for _, val := range dnsAndTrans.Spec.Paths {
			path2Svcs = append(path2Svcs, Path2Svc{Path: val.Name, Svc: val.Service})
		}
		bDnsAndTrans.Path2Svcs = path2Svcs
		results = append(results, bDnsAndTrans)
	}
	fmt.Print(DnsAndTransHeader())
	for _, v := range results {
		fmt.Print(v.ToString())
	}
}

func caseNode(name string) {
	url := baseUrl + path.Join(config.NODE_PREFIX, name)
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var results []*beautifiedNode
	for _, res := range listRes {
		node := &object.Node{}
		err = json.Unmarshal(res.ValueBytes, node)
		if err != nil {
			continue
		}
		bNode := &beautifiedNode{
			Name:          node.MetaData.Name,
			Ctime:         node.MetaData.Ctime,
			MasterIp:      node.MasterIp,
			NodeIp:        node.Spec.DynamicIp,
			NodeIpAndMask: node.Spec.NodeIpAndMask,
		}
		results = append(results, bNode)
	}
	fmt.Print(NODEHeader())
	for _, bNode := range results {
		fmt.Print(bNode.ToString())
	}
}

func casePod(name string) {
	url := baseUrl + path.Join(config.PodRuntimePrefix, name)
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var results []*beautifiedPod
	for _, res := range listRes {
		pod := &object.Pod{}
		err = json.Unmarshal(res.ValueBytes, pod)
		if err != nil {
			continue
		}
		bPod := &beautifiedPod{
			Name:     pod.Name,
			Ctime:    pod.Ctime,
			PodIp:    pod.Status.PodIP,
			NodeName: pod.Spec.NodeName,
			Status:   pod.Status.Phase,
		}
		results = append(results, bPod)
	}
	fmt.Print(PODHeader())
	for _, bPod := range results {
		fmt.Print(bPod.ToString())
	}
}
func caseService(name string) {
	url := baseUrl + path.Join(config.ServicePrefix, name)
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var results []*beautifiedService
	for _, res := range listRes {
		service := &object.Service{}
		err = json.Unmarshal(res.ValueBytes, service)
		if err != nil {
			continue
		}
		bService := &beautifiedService{
			Name:      service.MetaData.Name,
			Ctime:     service.MetaData.Ctime,
			ClusterIP: service.Spec.ClusterIp,
			Status:    service.Status.Phase,
		}
		var ports []PortAndProtocol
		for _, port := range service.Spec.Ports {
			ports = append(ports, PortAndProtocol{
				Port:     port.Port,
				Protocol: port.Protocol,
			})
		}
		bService.Ports = ports
		results = append(results, bService)
	}
	fmt.Print(SERVICEHeader())
	for _, bsvc := range results {
		fmt.Print(bsvc.ToString())
	}
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

func caseAutoscaler(name string) {
	url := baseUrl + path.Join("/registry/autoscaler/default", name)
	listRes, err := client.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	bas := beautifiedAutoscaler{}
	var bass []*beautifiedAutoscaler
	for _, res := range listRes {
		autoscaler := object.Autoscaler{}
		err := json.Unmarshal(res.ValueBytes, &autoscaler)
		if err != nil {
			continue
		}
		bas.name = autoscaler.Metadata.Name
		bas.refKind = autoscaler.Spec.ScaleTargetRef.Kind
		bas.refName = autoscaler.Spec.ScaleTargetRef.Name
		bas.maxReplicas = autoscaler.Spec.MaxReplicas
		bas.minReplicas = autoscaler.Spec.MinReplicas
		bas.scaleInterval = autoscaler.Spec.ScaleInterval
		bas.metrics = autoscaler.Spec.Metrics
		bass = append(bass, &bas)
	}
	fmt.Print(BASHeader())
	for _, i := range bass {
		fmt.Print(i.ToString())
	}
}

func caseJob(name string) {
	url := baseUrl + path.Join(config.Job2PodPrefix, name)
	if name == "" {
		listRes, err := client.Get(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		results := make(map[string]string)
		var tmp object.Job2Pod
		for _, res := range listRes {
			err = json.Unmarshal(res.ValueBytes, &tmp)
			if err != nil {
				continue
			}
			results[path.Base(res.Key)] = tmp.PodName
		}
		fmt.Printf("%-50s\t%-50s\n", "Job", "CommittedBy")
		for k, v := range results {
			fmt.Printf("%-50s\t%-50s\n", k, v)
		}
	} else {
		listRes, err := client.Get(url)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(listRes) == 0 {
			fmt.Println("Job not found")
		}
		var tmp object.Job2Pod
		err = json.Unmarshal(listRes[0].ValueBytes, &tmp)
		if err != nil {
			fmt.Println("Job not found")
			return
		}
		fmt.Printf("%-15s%-50s\n%-15s%-50s\n", "Job", path.Base(listRes[0].Key), "CommittedBy", tmp.PodName)

		listRes, err = client.Get(baseUrl + path.Join(config.PodRuntimePrefix, tmp.PodName))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(listRes) == 0 {
			fmt.Println("Pod runtime not found.")
			return
		}
		var pod object.Pod
		err = json.Unmarshal(listRes[0].ValueBytes, &pod)
		if err != nil {
			fmt.Println("[error] Cannot get pod IP.")
			return
		}
		fmt.Printf("Pod IP: %s. Check SharedData of the pod's host node to get the output.\nWaiting for remote server to respond, it takes a few time...\n\n", pod.Status.PodIP)
		data, err := client.RawGet("http://" + pod.Status.PodIP + ":9990")
		fmt.Println(string(data))
	}
}
