package main

import (
	"fmt"
	"minik8s/pkg/iptables"
	"os"
	"path"
	"path/filepath"
)

func GetProjectAbsPath() (projectAbsPath string) {
	programPath, _ := filepath.Abs(os.Args[0])
	fmt.Println("programPath:", programPath)
	projectAbsPath = path.Dir(path.Dir(programPath))
	fmt.Println("PROJECT_ABS_PATH:", projectAbsPath)
	return projectAbsPath
}
func main() {
	ipt, err := iptables.New()
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
	}

	//nodePort part
	//err = ipt.NewChain("nat", "HONG-NODE-PORT")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//err = ipt.Append("nat", "HONG-SERVICE", "-p", "all", "-j", "HONG-NODE-PORT", "-m", "addrtype", "--dst-type", "LOCAL")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//err = ipt.Append("nat", "HONG-NODE-PORT", "-p", "tcp", "--dport", "8070", "-j", "HONG-SVC")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//create HONG-SEP chain

	//err = ipt.NewChain("nat", "HONG-SEP")
	//if err != nil {
	//	fmt.Println(err)
	//}
	err = ipt.Append("nat", "HONG-SEP", "-s", "0/0", "-d", "0/0", "-p", "tcp", "-j", "DNAT", "--to-destination", "172.16.16.2:80")
	if err != nil {
		fmt.Println(err)
	}

	//create HONG-SVC chain
	//err = ipt.NewChain("nat", "HONG-SVC")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//err = ipt.Append("nat", "HONG-SVC", "-p", "tcp", "-j", "HONG-SEP")
	//if err != nil {
	//	fmt.Println(err)
	//}

	//create HONG-SERVICE chain
	//err = ipt.NewChain("nat", "HONG-SERVICE")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//err = ipt.Append("nat", "HONG-SERVICE", "-s", "0/0", "-d", "10.12.34.45", "-p", "tcp", "--dport", "8070", "-j", "HONG-SVC")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//////add into OUTPUT chain
	//err = ipt.Insert("nat", "PREROUTING", 1, "-j", "HONG-SERVICE", "-s", "0/0", "-d", "0/0", "-p", "all")
	//err = ipt.Insert("nat", "OUTPUT", 1, "-j", "HONG-SERVICE", "-s", "0/0", "-d", "0/0", "-p", "all")
	//if err != nil {
	//	fmt.Println(err)
	//}
}
