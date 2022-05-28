package mesh

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"strconv"
	"syscall"
)

var (
	BasePort  int64 = 16001
	RangePort int64 = 100
)

type Proxy struct {
	PodIP   string
	Address string
	server  *net.TCPListener
}

func NewProxy(podIP string) *Proxy {
	return &Proxy{
		PodIP: podIP,
	}
}

// select a spare port

// check & add iptables rules
func (p *Proxy) natRedirect() {

}

func (p *Proxy) Init() {
	var lnaddr *net.TCPAddr
	var server *net.TCPListener
	var err error

	for i := BasePort; i < BasePort+RangePort; i++ {
		lnaddr, err = net.ResolveTCPAddr("tcp", "127.0.0.1:"+strconv.Itoa(int(i)))
		if err != nil {
			continue
		}

		server, err = net.ListenTCP("tcp", lnaddr)
		if err == nil {
			p.Address = "127.0.0.1:" + strconv.Itoa(int(i))
			p.server = server
			break
		}
	}

	if err != nil || server == nil {
		fmt.Println("[Proxy] No port available")
		return
	}

	fmt.Printf("[Proxy] listening to:%v\n", p.Address)

	err = p.initChain()
	if err != nil {
		fmt.Printf("[FATAL] init iptables chain fail...\n")
		return
	}
}

func (p *Proxy) Run() {
	if p.server == nil {
		fmt.Println("[Proxy Run] No server available")
	}

	func(p *Proxy) {
		defer p.finalizeChain()

		for {
			conn, err := p.server.AcceptTCP()
			if err != nil {
				continue
			}
			go handleConn(conn)
		}
	}(p)

}

func handleConn(clientConn *net.TCPConn) {

	fmt.Printf("connection from:%v...\n", clientConn.RemoteAddr().String())

	if clientConn == nil {
		return
	}

	ipv4, port, clientConn, err := getOriginalDst(clientConn)

	if err != nil {
		return
	}

	fmt.Printf("To %v:%v", ipv4, port)

	// TODO: clusterIP to a endpoint
	// by ip or by regex

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

	defer dst.Close()
	defer src.Close()

	_, err := io.Copy(dst, src)
	if err != nil {
		fmt.Println("[copy] error")
		return
	}
}
