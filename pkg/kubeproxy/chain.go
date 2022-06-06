package kubeproxy

import (
	"fmt"
	"minik8s/pkg/iptables"
)

type DnatRule struct {
	RulesSpec []string
	//pod的远程地址
	PodIp string
	//pod端口
	Port string
	//协议类型
	Protocol string
	//所属表
	Table string
	//所属链
	FatherChain string
}

func (rule *DnatRule) formRuleSpec() {
	tmp := []string{"-s", "0/0", "-d", "0/0"}
	tmp = append(tmp, "-p", rule.Protocol, "-j", "DNAT", "--to-destination", rule.PodIp+":"+rule.Port)
	rule.RulesSpec = tmp
}

func NewDnatRule(podIp string, port string, protocol string, table string, fatherChain string) *DnatRule {
	res := &DnatRule{
		PodIp:       podIp,
		Port:        port,
		Protocol:    protocol,
		Table:       table,
		FatherChain: fatherChain,
	}
	res.formRuleSpec()
	return res
}
func (rule *DnatRule) ApplyRule() error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	err = ipt.Append(rule.Table, rule.FatherChain, rule.RulesSpec...)
	if err != nil {
		return err
	}
	return nil
}
func (rule *DnatRule) DeleteRule() error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	err = ipt.Delete(rule.Table, rule.FatherChain, rule.RulesSpec...)
	if err != nil {
		return err
	}
	return nil
}

//SEP链，一条链对应一个pod, 对应一条Dnat规则
type PodUnit struct {
	PodIp   string
	PodName string
	PodPort string
}
type SepChain struct {
	//链的名字
	Name string
	//协议类型
	Protocol string
	//所属表
	Table string
	//所属链
	FatherChain string
	//PodIp
	PodIp string
	//PodName
	PodName string
	//PodPort
	PodPort string
	//在父链中应用该链的规则
	RuleSpec []string
	//包含的Dnat rule
	DNatRule *DnatRule
	//Round Robin num, 每n个包执行该规则
	RrNum int
}

func (chain *SepChain) formRuleSpec() {
	tmp := []string{"-p", chain.Protocol, "-m", "statistic", "--mode", "nth", "--every", fmt.Sprintf("%d", chain.RrNum),
		"--packet", "0", "-j", chain.Name}
	chain.RuleSpec = tmp
}

//新建一条SepChain

func NewSepChain(protocol string, table string, fatherChain string, podIp string, podName string, podPort string, rrNum int) *SepChain {
	res := &SepChain{
		Name:        SepChainPrefix + "-" + podName + podPort,
		Protocol:    protocol,
		Table:       table,
		FatherChain: fatherChain,
		PodIp:       podIp,
		PodName:     podName,
		PodPort:     podPort,
		RrNum:       rrNum,
	}
	res.formRuleSpec()
	//创建该链
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("[chain] NewSepChain Error")
		fmt.Println(err)
	}
	err = ipt.NewChain(table, res.Name)
	if err != nil {
		fmt.Println("[chain] NewSepChain Error")
		fmt.Println(err)
	}
	//为Sep链增加DNat规则
	res.DNatRule = NewDnatRule(podIp, podPort, protocol, table, res.Name)
	err = res.DNatRule.ApplyRule()
	if err != nil {
		fmt.Println("[chain] NewSepChain Error")
		fmt.Println(err)
	}
	return res
}
func (chain *SepChain) ApplyRule() error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	err = ipt.Append(chain.Table, chain.FatherChain, chain.RuleSpec...)
	if err != nil {
		return err
	}
	return nil
}
func (chain *SepChain) DeleteRule() error {
	//删除该链，所有相关的都要清除干净
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	//先删除在父链中的规则
	err = ipt.Delete(chain.Table, chain.FatherChain, chain.RuleSpec...)
	if err != nil {
		return err
	}
	//删除sep链自身中的DNat规则
	err = chain.DNatRule.DeleteRule()
	if err != nil {
		return err
	}
	//删除该链
	err = ipt.DeleteChain(chain.Table, chain.Name)
	return err
}

type SvcChain struct {
	//链名
	Name        string
	Table       string
	FatherChain string
	//clusterIp
	ClusterIp string
	//clusterPort
	ClusterPort string
	Protocol    string
	//podName到SepChain的映射
	PodName2SepChain map[string]*SepChain
	//应用链的规则
	RuleSpec []string
}

func NewSvcChain(serviceName string, table string, fatherChain string, clusterIp string, clusterPort string, protocol string, units []PodUnit) *SvcChain {
	res := &SvcChain{
		Name:        SvcChainPrefix + "-" + serviceName + clusterPort,
		Table:       table,
		FatherChain: fatherChain,
		ClusterPort: clusterPort,
		ClusterIp:   clusterIp,
		Protocol:    protocol,
	}
	res.formRuleSpec()
	//创建该链
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("[chain] NewSvcChain Error")
		fmt.Println(err)
	}
	err = ipt.NewChain(table, res.Name)
	//生成对应的sep链，并apply
	total := len(units)
	res.PodName2SepChain = make(map[string]*SepChain)
	for _, val := range units {
		sepChain := NewSepChain(protocol, table, res.Name, val.PodIp, val.PodName, val.PodPort, total)
		err = sepChain.ApplyRule()
		if err != nil {
			fmt.Println("[chain] NewSvcChain Error")
			fmt.Println(err)
		}
		total--
		res.PodName2SepChain[val.PodName] = sepChain
	}
	return res
}
func (chain *SvcChain) formRuleSpec() {
	tmp := []string{"-s", "0/0", "-d", chain.ClusterIp, "-p", chain.Protocol, "--dport", chain.ClusterPort, "-j", chain.Name}
	chain.RuleSpec = tmp
}

func (chain *SvcChain) ApplyRule() error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	err = ipt.Append(chain.Table, chain.FatherChain, chain.RuleSpec...)
	if err != nil {
		return err
	}
	return nil
}
func (chain *SvcChain) DeleteRule() error {
	ipt, err := iptables.New()
	if err != nil {
		return err
	}
	//删除在父链中的该规则
	err = ipt.Delete(chain.Table, chain.FatherChain, chain.RuleSpec...)
	if err != nil {
		return err
	}
	//删除该链下的所有sep链
	for _, ch := range chain.PodName2SepChain {
		err = ch.DeleteRule()
		if err != nil {
			return err
		}
	}
	//删除该链
	err = ipt.DeleteChain(chain.Table, chain.Name)
	return err
}

func (chain *SvcChain) UpdateRule(newUnits []PodUnit) {
	//根据新的podUnit来调整该SVC链下的SEP链，注意需要调整RR中的参数
	//先把消失的删了
	remain := make(map[string]*SepChain)
	var err error
	for k, v := range chain.PodName2SepChain {
		isRemain := false
		for _, newUnit := range newUnits {
			if newUnit.PodName == k {
				isRemain = true
				break
			}
		}
		if isRemain {
			remain[k] = v
		} else {
			err = v.DeleteRule()
			if err != nil {
				fmt.Println("[chain] UpdateRule error")
				fmt.Println(err)
			}
		}
	}
	chain.PodName2SepChain = make(map[string]*SepChain)
	//进行更新以及建立新的
	//先把旧的规则全删了
	ipt, _ := iptables.New()
	for _, v := range remain {
		err = ipt.Delete(v.Table, v.FatherChain, v.RuleSpec...)
		if err != nil {
			fmt.Println("[chain] UpdateRule error")
			fmt.Println(err)
		}
	}
	total := len(newUnits)
	for _, newUnit := range newUnits {
		sepChain, ok := remain[newUnit.PodName]
		if ok {
			//存在sep链，更新即可
			sepChain.RrNum = total
			sepChain.formRuleSpec()
			err = ipt.Append(sepChain.Table, sepChain.FatherChain, sepChain.RuleSpec...)
			if err != nil {
				fmt.Println("[chain] UpdateRule error")
				fmt.Println(err)
			}
			chain.PodName2SepChain[newUnit.PodName] = sepChain
		} else {
			//需要创建新的
			sepChain = NewSepChain(chain.Protocol, chain.Table, chain.Name, newUnit.PodIp, newUnit.PodName, newUnit.PodPort, total)
			err = sepChain.ApplyRule()
			if err != nil {
				fmt.Println("[chain] UpdateRule error")
				fmt.Println(err)
			}
			chain.PodName2SepChain[newUnit.PodName] = sepChain
		}
		total--
	}
}

//启动的函数,用于创建services链并加入到OUTPUT以及PREROUTING链中

func Boot() {
	//先判断services链存不存在
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	exist, err2 := ipt.ChainExists(NatTable, GeneralServiceChain)
	if err2 != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	if exist {
		return
	}
	//创建该链并做处理
	err = ipt.NewChain(NatTable, GeneralServiceChain)
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	err = ipt.Insert(NatTable, OutPutChain, 1, "-j", GeneralServiceChain, "-s", "0/0", "-d", "0/0", "-p", "all")
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
	err = ipt.Insert(NatTable, PreRoutingChain, 1, "-j", GeneralServiceChain, "-s", "0/0", "-d", "0/0", "-p", "all")
	if err != nil {
		fmt.Println("[chain] Boot error")
		fmt.Println(err)
	}
}
