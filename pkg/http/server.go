package httpserver

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
)

type Server struct {
	Addr        string
	Handler     http.Handler
	socket      *TcpSocket
	connections map[int]bool
}

func (s *Server) ListenAndServeHttp() error {
	s.connections = map[int]bool{}

	parseAddr := strings.Split(s.Addr, ":")
	if len(parseAddr) != 2 {
		return errors.New("server Addr requires full ip address and port, eg 0.0.0.0:8080")
	}

	addr := parseAddr[0]
	portStr := parseAddr[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	netIp, err := parseIP(addr)
	if err != nil {
		return err
	}

	ip := [4]byte{}
	copy(ip[:], netIp)

	tcp, err := newSocket(port, ip)
	if err != nil {
		return err
	}

	s.socket = tcp

	err = s.socket.Listen()
	if err != nil {
		return err
	}

	for {
		conn, err := s.socket.Accept()
		if err != nil {
			if err.Error() == "software caused connection abort" {
				break
			}

			return err
		}
		fd := conn.FD()
		s.connections[fd] = true

		go func(c *Conn) error {
			defer func() {
				c.Close()
				delete(s.connections, c.FD())
			}()
			r, err := HandleRequest(c)
			if err != nil {
				return err
			}

			w := NewResponseWriter()
			w.Version(r.Proto)

			s.Handler.ServeHTTP(w, r)

			n, _ := conn.Write(w.Marshal())

			return nil
		}(conn)
	}

	return nil
}

func (s *Server) Close() error {
	if s.socket == nil {
		return errors.New("socket not opened")
	}

	for k, v := range s.connections {
		if v {
			syscall.Close(k)
		}
	}

	return s.socket.Close()
}

func newSocket(port int, addr [4]byte) (*TcpSocket, error) {
	sa := &syscall.SockaddrInet4{
		Port: 8080,
		Addr: addr,
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}

	err = syscall.SetNonblock(fd, false)
	if err != nil {
		return nil, syscall.Close(fd)
	}

	err = syscall.Bind(fd, sa)
	if err != nil {
		return nil, syscall.Close(fd)
	}

	s := &TcpSocket{
		fd:   fd,
		port: port,
		addr: addr,
	}

	return s, nil
}

func parseIP(addr string) (net.IP, error) {
	if len(addr) == 0 {
		return nil, errors.New("no address provided")
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, errors.New("invalid address string " + addr)
	}

	return ip, nil
}

func parseHttpRequest(data []byte, conn *Conn) (*http.Request, error) {
	rdr := bytes.NewReader(data)
	scanner := bufio.NewScanner(rdr)

	scanner.Scan()
	requestLine := scanner.Text()
	if requestLine == "" {
		return nil, errors.New("no valid request line")
	}
	requestLineFields := strings.Fields(requestLine)

	headers := map[string][]string{}
	contentLength := 0
	host := ""

	for scanner.Scan() {
		next := scanner.Text()
		if next == "" {
			break
		}

		fields := strings.Fields(next)
		var f string
		if len(fields) > 0 {
			f = fields[0]
			headers[f] = []string{}
		} else {
			continue
		}

		if len(fields) > 1 {
			values := strings.Fields(fields[1])
			headers[f] = append(headers[f], values...)

			if f == "Content-Length" {
				contentLength, _ = strconv.Atoi(values[0])
			}

			if f == "host" {
				host = values[0]
			}
		}

	}

	// build body after double line break
	body := []byte{}
	for scanner.Scan() {
		b := scanner.Bytes()
		body = append(body, b...)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	bodyReader := bytes.NewReader(body)

	readCloser := io.NopCloser(bodyReader)
	defer readCloser.Close()

	// build request
	remoteAddr, _ := conn.RemoteAddr()

	req := &http.Request{
		Method: requestLineFields[0],
		Proto:  requestLineFields[2],
		URL: &url.URL{
			Path: requestLineFields[1],
		},
		Header:        headers,
		Body:          readCloser,
		ContentLength: int64(contentLength),
		Host:          host,
		RequestURI:    requestLineFields[1],
		RemoteAddr:    remoteAddr,
	}

	return req, nil
}
