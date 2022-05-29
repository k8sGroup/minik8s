package service

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/netSupport/netconfig"
)

var gateWayRsModule *object.ReplicaSet
var coreDnsRsModule *object.ReplicaSet
var coreDnsServiceModule *object.Service
var gateWayServiceModule *object.Service

func getGateWayRsModule() *object.ReplicaSet {
	if gateWayRsModule == nil {
		data, err := ioutil.ReadFile(netconfig.GateWayRsModulePath)
		if err != nil {
			fmt.Println("[dnsModule] getGateWayRsModule fail" + err.Error())
			return nil
		}
		err = yaml.Unmarshal([]byte(data), gateWayRsModule)
		if err != nil {
			fmt.Println("[dnsModule] getGateWayRsModule fail" + err.Error())
			return nil
		}
		return gateWayRsModule
	} else {
		return gateWayRsModule
	}
}
func getCoreDnsRsModule() *object.ReplicaSet {
	if coreDnsRsModule == nil {
		data, err := ioutil.ReadFile(netconfig.CoreDnsRsModulePath)
		if err != nil {
			fmt.Println("[dnsModule] GetCoreDnsRsModule fail" + err.Error())
			return nil
		}
		err = yaml.Unmarshal([]byte(data), coreDnsRsModule)
		if err != nil {
			fmt.Println("[dnsModule] GetCoreDnsRsModule fail" + err.Error())
			return nil
		}
		return coreDnsRsModule
	} else {
		return coreDnsRsModule
	}
}
func getGateWayServiceModule() *object.Service {
	if gateWayServiceModule == nil {
		data, err := ioutil.ReadFile(netconfig.GateWayServiceModulePath)
		if err != nil {
			fmt.Println("[dnsModule] getGateWayServiceModule fail" + err.Error())
			return nil
		}
		err = yaml.Unmarshal([]byte(data), gateWayServiceModule)
		if err != nil {
			fmt.Println("[dnsModule] getGateWayServiceModule fail" + err.Error())
			return nil
		}
		return gateWayServiceModule
	} else {
		return gateWayServiceModule
	}
}
func getCoreDnsServiceModule() *object.Service {
	if coreDnsServiceModule == nil {
		data, err := ioutil.ReadFile(netconfig.CoreDnsServiceModulePath)
		if err != nil {
			fmt.Println("[dnsModule] getCoreDnsServiceModule fail" + err.Error())
			return nil
		}
		err = yaml.Unmarshal([]byte(data), coreDnsServiceModule)
		if err != nil {
			fmt.Println("[dnsModule] getCoreDnsServiceModule fail" + err.Error())
			return nil
		}
		return coreDnsServiceModule
	} else {
		return coreDnsServiceModule
	}
}

func GetGateWayRsModule(DnsAndTransName string) *object.ReplicaSet {
	origin := getGateWayRsModule()
	newSpec := origin.Spec
	newSpec.Template.Labels[netconfig.BelongKey] = DnsAndTransName
	newSpec.Template.Name = netconfig.GateWayPodNamePrefix + DnsAndTransName
	newSpec.Template.Spec.Volumes[0].Path = netconfig.NginxPathPrefix + "/" + DnsAndTransName
	newSpec.Template.Spec.Containers[0].Name = netconfig.GateWayContainerPrefix + DnsAndTransName
	res := &object.ReplicaSet{
		object.ObjectMeta{
			Name:   netconfig.GateWayRsNamePrefix + DnsAndTransName,
			Labels: origin.Labels,
		},
		newSpec,
		origin.Status,
	}
	return res
}
func GetCoreDnsRsModule() *object.ReplicaSet {
	origin := getCoreDnsRsModule()
	res := &object.ReplicaSet{
		origin.ObjectMeta,
		origin.Spec,
		origin.Status,
	}
	return res
}
func GetGateWayServiceModule(DnsAndTransName string) *object.Service {
	origin := getGateWayServiceModule()
	meta := origin.MetaData
	meta.Name = netconfig.GateWayServicePrefix + DnsAndTransName
	spec := origin.Spec
	spec.Selector[netconfig.BelongKey] = DnsAndTransName
	res := &object.Service{meta, spec, origin.Status}
	return res
}
func GetCoreDnsServiceModule() *object.Service {
	origin := getCoreDnsServiceModule()
	res := &object.Service{
		origin.MetaData,
		origin.Spec,
		origin.Status,
	}
	return res
}
