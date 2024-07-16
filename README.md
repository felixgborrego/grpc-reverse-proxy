# gRPC Reverse Proxy

This is a simple gRPC reverse proxy that can be used as a sidecar to proxy gRPC requests to a remote server.

# Why this project exists?

While there are many alternatives for sidecar solutions, such as service meshes based on Istio or more advanced approaches like eBPF with Cilium, sometimes it's just easier to understand a simple code than a complex framework. This project aims to provide a straightforward gRPC reverse proxy that can be easily understood and deployed without the overhead of learning and managing a full-fledged service mesh or advanced networking tools. 

By focusing on simplicity, this project allows developers to quickly set up a reverse proxy for gRPC services, making it an ideal choice for smaller projects or for those who prefer minimalistic solutions.


# Usage

Typicall this will be deployed as a docker conatiner, but you can run it locally with:

```
# Example to proxy a pubsub service.
export PROXY_TARGET=europe-west1-pubsub.googleapis.com
export PROXY_AUTHORITY=europe-west1-pubsub.googleapis.com
export TLS_CERT_FILE=sidecar.pem
export TLS_KEY_FILE=sidecar.key
go run .
```

Note you can generate a local pem and key with:
```
openssl req -x509 -newkey rsa:4096 -keyout sidecar.key -out sidecar.pem -days 365 -nodes -subj "/CN=localhost"

