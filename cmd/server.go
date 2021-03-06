package main

import (
	"context"
	"log"
	"net"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	rpc "gitlab.com/whom/rpc/rpc"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	rpc.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *rpc.HelloRequest) (*rpc.HelloReply, error) {
	p, ok := peer.FromContext(ctx)
	if ok {
		tlsInfo := p.AuthInfo.(credentials.TLSInfo)
		clientCert := tlsInfo.State.VerifiedChains[0][0].Subject
		clientCaCert := tlsInfo.State.VerifiedChains[0][1].Subject
		log.Printf("Client Cert: %s", clientCert)
		log.Printf("Client CA Cert: %s", clientCaCert)
	}

	log.Printf("Received: %v", in.GetName())
	return &rpc.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	certificate, cerr := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if cerr != nil {
		log.Fatal("cannot load server keypair: ", cerr)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatal("could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append client certs")
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	s := grpc.NewServer(
		grpc.Creds(creds),
	)

	rpc.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
