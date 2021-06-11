package main

import (
	"context"
	"log"
	"os"
	"fmt"
	"time"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	rpc "gitlab.com/whom/rpc/rpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func loadTLSCredentials() (credentials.TransportCredentials, error) {
    caCert, err := ioutil.ReadFile("certs/ca.crt")
    if err != nil {
        return nil, err
    }

    certPool := x509.NewCertPool()
    if !certPool.AppendCertsFromPEM(caCert) {
        return nil, fmt.Errorf("failed to add server CA's certificate")
    }

    // Create the credentials and return it
    config := &tls.Config{
        RootCAs:      certPool,
    }

    return credentials.NewTLS(config), nil
}


func main() {
	certificate, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatalf("could not load client key pair: %s", err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   "localhost",
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	conn, derr := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if derr != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := rpc.NewGreeterClient(conn)


	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &rpc.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
