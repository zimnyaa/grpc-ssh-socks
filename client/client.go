package main

import (
	"golang.org/x/crypto/ssh"
	"net"
	"io"
	"log"
	"context"
	"fmt"
	"crypto/rand"
	"crypto/rsa"
	"zimnyaa/grpc-ssh-socks/share"
	"google.golang.org/grpc"
	"zimnyaa/grpc-ssh-socks/grpctun"
)


func main() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	grpcconn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer grpcconn.Close()


	client := grpctun.NewTunnelServiceClient(grpcconn)
	stream, err := client.Tunnel(context.Background())
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}
	nConn := share.NewGrpcClientConn(stream)


	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		panic("Failed to create signer")
	}

	config.AddHostKey(signer)

	sshConn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}
	defer sshConn.Close()


	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		fmt.Println("new channel")
		if newChannel.ChannelType() != "direct-tcpip" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		fmt.Println("new direct-tcpip channel")
		

		var dReq struct {
			DestAddr string
			DestPort uint32
		}
		ssh.Unmarshal(newChannel.ExtraData(), &dReq)

		go func() {
			dest := fmt.Sprintf("%s:%d", dReq.DestAddr, dReq.DestPort)
			conn, err := net.Dial("tcp", dest)
				
			if err == nil {
				channel, chreqs, _ := newChannel.Accept()
				go ssh.DiscardRequests(chreqs)
	
				go func() {
					defer channel.Close()
					defer conn.Close()
					io.Copy(channel, conn)
				}()
				go func() {
					defer channel.Close()
					defer conn.Close()
					io.Copy(conn, channel)
				}()
			}
		}()
	}
	
}
