package mesh

const (
	PreRoutingChain string = "PREROUTING"
	NatTable        string = "nat"
	TCP             string = "tcp"
)

//
//func (p *Proxy) initChain() error {
//	ipt, err := iptables.New()
//	if err != nil {
//		fmt.Printf("[initChain] new iptables error:%v\n", err)
//		return err
//	}
//
//	exist, err := ipt.ChainExists(NatTable, PreRoutingChain)
//	if err != nil || !exist {
//		fmt.Printf("[initChain] prerouting may not exist error:%v\n", err)
//		return errors.New("[initChain] chain not exist")
//	}
//
//	// redirect output network
//	exist, err = ipt.Exists(NatTable, PreRoutingChain, "-p", TCP, "-s", p.PodIP, "-j", "DNAT", "--to-destination", p.Address)
//	if err != nil {
//		fmt.Printf("[initChain] output rule exist checking error:%v\n", err)
//		return err
//	}
//	if !exist {
//		err = ipt.Insert(NatTable, PreRoutingChain, 1, "-p", TCP, "-s", p.PodIP, "-j", "DNAT", "--to-destination", p.Address)
//		if err != nil {
//			fmt.Printf("[initChain] output rule insert error:%v\n", err)
//			return err
//		}
//	} else {
//		fmt.Printf("[initChain] output rule already exist\n")
//	}
//
//	// redirect input network
//	exist, err = ipt.Exists(NatTable, PreRoutingChain, "-p", TCP, "-d", p.PodIP, "-j", "DNAT", "--to-destination", p.Address)
//	if err != nil {
//		fmt.Printf("[initChain] input rule exist checking error:%v\n", err)
//		return err
//	}
//	if !exist {
//		err = ipt.Insert(NatTable, PreRoutingChain, 1, "-p", TCP, "-d", p.PodIP, "-j", "DNAT", "--to-destination", p.Address)
//		if err != nil {
//			fmt.Printf("[initChain] input rule insert error:%v\n", err)
//			return err
//		}
//	} else {
//		fmt.Printf("[initChain] input rule already exist\n")
//	}
//	return nil
//}
//
//func (p *Proxy) finalizeChain() {
//	fmt.Printf("[finalizeChain] exit...\n")
//
//	ipt, err := iptables.New()
//	if err != nil {
//		fmt.Printf("[finalizeChain] new iptables error:%v\n", err)
//		return
//	}
//
//	err = ipt.DeleteIfExists(NatTable, PreRoutingChain, "-p", TCP, "-s", p.PodIP, "-j", "DNAT", "--to-destination", p.Address)
//	if err != nil {
//		fmt.Printf("[finalizeChain] output rule delete error:%v\n", err)
//	}
//
//	err = ipt.DeleteIfExists(NatTable, PreRoutingChain, "-p", TCP, "-d", p.PodIP, "-j", "DNAT", "--to-destination", p.Address)
//	if err != nil {
//		fmt.Printf("[finalizeChain] input rule delete error:%v\n", err)
//	}
//}
