package httpserver

import "syscall"

type TcpSocket struct {
	addr [4]byte
	fd   int
	port int
}

func (ts *TcpSocket) Listen() error {
	err := syscall.Listen(ts.fd, 20)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TcpSocket) Accept() (*Conn, error) {
	nfd, nsa, err := syscall.Accept(ts.fd)
	if err != nil {
		return nil, err
	}

	return &Conn{fd: nfd, sa: nsa}, nil
}

func (ts *TcpSocket) Close() error {
	return syscall.Close(ts.fd)
}

type Conn struct {
	fd int
	sa syscall.Sockaddr
}

func (c *Conn) Read(b []byte) (int, error) {
	return syscall.Read(c.fd, b)
}

func (c *Conn) Write(p []byte) (int, error) {
	return syscall.Write(c.fd, p)
}

func (c *Conn) Close() error {
	return syscall.Close(c.fd)
}

func (c *Conn) RemoteAddr() (string, error) {
	sa, err := syscall.Getpeername(c.fd)
	if err != nil {
		return "", err
	}

	addr := ""
	switch sa.(type) {
	case *syscall.SockaddrInet4:
		ip4, _ := sa.(*syscall.SockaddrInet4)
		ip := [4]byte{}
		copy(ip[:], ip4.Addr[:])
		for _, b := range ip {
			addr += string(b)
			addr += "."
		}
	default:
		return "", nil
	}

	return addr, nil
}

func (c *Conn) FD() int {
	return c.fd
}
