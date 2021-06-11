#!/bin/bash

# gen ca
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -sha256 -days 1825 -out ca.crt -config ca.conf

# gen srv cert
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 825 -sha256 -extensions req_ext -extfile server.conf

# gen client cert
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -config client.conf
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 825 -sha256 -extensions req_ext -extfile client.conf
