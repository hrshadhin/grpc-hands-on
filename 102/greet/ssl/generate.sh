#!/bin/bash

SERVER_CN=localhost

## Remove old files
rm ca.crt ca.key server.crt  server.csr server.key server.pem

## create Certificate Authority Key
openssl genrsa -passout pass:1111 -des3 -out ca.key 4096

## create Certificate Authority Trust certificate
openssl req -passin pass:1111 -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=${SERVER_CN}"

echo "CA's self-signed certificate"
openssl x509 -in ca.crt -noout -text

## create server key
openssl genrsa -passout pass:1111 -des3 -out server.key 4096
## convert server key to pem file
openssl pkcs8 -topk8 -nocrypt -passin pass:1111 -in server.key -out server.pem

## create server certificate signing request
openssl req -passin pass:1111 -new -key server.key -out server.csr -subj "/CN=${SERVER_CN}"
## create server certificate
openssl x509 -req -passin pass:1111 -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server.crt -noout -text
