package main

import (
	"golang.org/x/crypto/ssh"
	"net"
	"fmt"
	"context"
	"github.com/armon/go-socks5"
	"zimnyaa/grpc-ssh-socks/share"
	"google.golang.org/grpc"
	"zimnyaa/grpc-ssh-socks/grpctun"
)

type server struct{
	grpctun.UnimplementedTunnelServiceServer
}

func (s *server) Tunnel(stream grpctun.TunnelService_TunnelServer) error {
	fmt.Println("new tunnel")
	socksconn := share.NewGrpcServerConn(stream)

	sshConf := &ssh.ClientConfig{
		User:            "e",
		Auth:            []ssh.AuthMethod{ssh.Password("e")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	
	c, chans, reqs, err := ssh.NewClientConn(socksconn, "255.255.255.255", sshConf)
	if err != nil {
		fmt.Println("%v", err)
		return err
	}
	sshConn := ssh.NewClient(c, chans, reqs)
	
	defer sshConn.Close()

	fmt.Println("connected to backwards ssh server")


	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sshConn.Dial(network, addr)
		},
	}

	serverSocks, err := socks5.New(conf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("creating a socks server")
	if err := serverSocks.ListenAndServe("tcp", "127.0.0.1:1080"); err != nil {
		fmt.Println("failed to create socks5 server", err)
	}

	return nil

}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Println("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	grpctun.RegisterTunnelServiceServer(s, &server{})
	s.Serve(lis)
}

