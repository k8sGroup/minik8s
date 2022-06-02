package mesh

import (
	"fmt"
	"io"
	"minik8s/pkg/listerwatcher"
	"net"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
)

var (
	InIP    = "0.0.0.0"
	InPort  = 15006
	OutIP   = "127.0.0.1"
	OutPort = 15001
)

type Address struct {
	IP   string
	Port int
}

type Proxy struct {
	router *Router
}

func NewProxy(lsConfig *listerwatcher.Config) *Proxy {
	return &Proxy{
		router: NewRouter(lsConfig),
	}
}

func (p *Proxy) Init(addresses []Address) {
	var lnaddr *net.TCPAddr
	var err error

	for _, addr := range addresses {
		lnaddr, err = net.ResolveTCPAddr("tcp", addr.IP+":"+strconv.Itoa(addr.Port))
		if err != nil {
			fmt.Println("[Proxy] No port available")
			return
		}

		server, err := net.ListenTCP("tcp", lnaddr)
		if err != nil {
			fmt.Println("[Proxy] init server fail")
			return
		}

		fmt.Println("[Init] listening to " + addr.IP + ":" + strconv.Itoa(addr.Port))

		go p.run(server)
	}

	go p.router.Run()
}

func (p *Proxy) run(server *net.TCPListener) {
	if server == nil {
		fmt.Println("[Proxy Run] No server available")
	}
	fmt.Println("Proxy Run...")
	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			continue
		}
		go p.handleConn(conn)
	}

}

func (p *Proxy) handleConn(clientConn *net.TCPConn) {

	fmt.Printf("connection from:%v...\n", clientConn.RemoteAddr().String())

	if clientConn == nil {
		return
	}

	ipv4, port, clientConn, err := getOriginalDst(clientConn)

	if err != nil {
		return
	}

	fmt.Printf("To %v:%v", ipv4, port)

	// clusterIP to an endpoint
	// if ipv4 is not clusterIP, endPoint will still be ipv4
	endpointIP, err := p.router.GetEndPoint(ipv4)
	if err != nil || endpointIP == nil {
		fmt.Printf("[handleConn] no endpoints for %v err:%v", ipv4, endpointIP)
		return
	}

	directConn, err := dial(*endpointIP, int(port))
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
