package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

func main() {

	lnaddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:16001")
	if err != nil {
		panic(err)
	}

	server, err := net.ListenTCP("tcp", lnaddr)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(clientConn *net.TCPConn) {

	fmt.Println("connect...")
	fmt.Println(clientConn.RemoteAddr().String())

	if clientConn == nil {
		return
	}

	ipv4, port, clientConn, err := getOriginalDst(clientConn)

	if err != nil {
		return
	}

	fmt.Printf("To %v:%v", ipv4, port)

	directConn, err := dial(ipv4, int(port))
	if err != nil {
		fmt.Printf("Could not connect, giving up: %v", err)
		return
	}
	fmt.Printf("Connected to remote end %v %v", clientConn.RemoteAddr(), directConn.RemoteAddr())

	go copy(clientConn, directConn)
	go copy(directConn, clientConn)
}

func getOriginalDst(clientConn *net.TCPConn) (ipv4 string, port uint16, newTCPConn *net.TCPConn, err error) {

	remoteAddr := clientConn.RemoteAddr()
	if remoteAddr == nil {
		err = fmt.Errorf("clientConn.fd is nil")
		return
	}

	newTCPConn = nil

	clientConnFile, err := clientConn.File()
	if err != nil {
		return
	} else {
		clientConn.Close()
	}

	addr, err := syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IP, 80)
	if err != nil {
		return
	}
	newConn, err := net.FileConn(clientConnFile)
	if err != nil {
		return
	}
	if _, ok := newConn.(*net.TCPConn); ok {
		newTCPConn = newConn.(*net.TCPConn)
		clientConnFile.Close()
	} else {
		errmsg := fmt.Sprintf("ERR: newConn is not a *net.TCPConn, instead it is: %T (%v)", newConn, newConn)
		err = errors.New(errmsg)
		return
	}

	ipv4 = itod(uint(addr.Multiaddr[4])) + "." +
		itod(uint(addr.Multiaddr[5])) + "." +
		itod(uint(addr.Multiaddr[6])) + "." +
		itod(uint(addr.Multiaddr[7]))
	port = uint16(addr.Multiaddr[2])<<8 + uint16(addr.Multiaddr[3])

	return
}

func itod(i uint) string {
	if i == 0 {
		return "0"
	}

	var b [32]byte
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}
	return string(b[bp:])
}

func dial(host string, port int) (*net.TCPConn, error) {
	remoteAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}
	remoteAddrAndPort := &net.TCPAddr{IP: remoteAddr.IP, Port: port}
	var localAddr *net.TCPAddr
	localAddr = nil
	conn, err := net.DialTCP("tcp", localAddr, remoteAddrAndPort)
	return conn, err
}

func copy(dst io.ReadWriteCloser, src io.ReadWriteCloser) {
	if dst == nil || src == nil {
		fmt.Println("[copy] null src/dst")
		return
	}
	_, err := io.Copy(dst, src)
	if err != nil {
		fmt.Println("[copy] error")
		return
	}

	dst.Close()
	src.Close()
}
