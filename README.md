# clamav-api-go

[![build and test](https://github.com/lescactus/clamav-api-go/actions/workflows/go.yaml/badge.svg)](https://github.com/lescactus/clamav-api-go/actions/workflows/go.yaml) [![Kubernetes](https://github.com/lescactus/clamav-api-go/actions/workflows/k8s.yaml/badge.svg)](https://github.com/lescactus/clamav-api-go/actions/workflows/k8s.yaml) [![Release](https://github.com/lescactus/clamav-api-go/actions/workflows/release.yml/badge.svg)](https://github.com/lescactus/clamav-api-go/actions/workflows/release.yml)

 Simple REST API wrapper for [ClamAV](http://www.clamav.net/) written in Go.

## Description

This is a REST API wrapper server with support for basic `INSTREAM` scanning, `VERSIONCOMMANDS`, `STATS`, `SHUTDOWN`, `VERSION`, `RELOAD` and `PING` command. 

The Clamd tcp protocol is explained here: http://linux.die.net/man/8/clamd

## Requirements

* Golang 1.17 or higher

## Getting started

### Building `clamav-api-go` :cd:

<details>

#### From source with Go

You need a working [go](https://golang.org/doc/install) toolchain (It has been developped and tested with go 1.20 and should work with go >= 1.20). Refer to the official documentation for more information (or from your Linux/Mac/Windows distribution documentation to install it from your favorite package manager).

```bash
# Clone this repository
git clone https://github.com/lescactus/clamav-api-go.git && cd clamav-api-go/

# Build from sources. Use the '-o' flag to change the compiled binary name
go build

# Default compiled binary is clamav-api-go
# You can optionnaly move it somewhere in your $PATH to access it shell wide
./clamav-api-go
```

#### From source with docker

If you don't have [go](https://golang.org/doc/install) installed but have docker, run the following command to build inside a docker container:

```bash
# Build from sources inside a docker container. Use the '-o' flag to change the compiled binary name
# Warning: the compiled binary belongs to root:root
docker run --rm -it -v "$PWD":/app -w /app golang:1.20 go build -buildvcs=false

# Default compiled binary is clamav-api-go
# You can optionnaly move it somewhere in your $PATH to access it shell wide
./clamav-api-go
```

The server is accessible at http://127.0.0.1:8080

#### With Docker

`clamav-api-go` comes with a `Dockerfile`. To build the image:

```bash
docker build -t clamav-api .

docker run -d -p 8080:8080 --restart="always" --name clamav-api-go clamav-api-go 
```

The server is accessible at http://127.0.0.1:8080

</details>

### Running with Docker :rooster:

```bash
docker run -d -p 8080:8080 --restart="always" --name clamav-api-go ghcr.io/lescactus/clamav-api-go
```

The server is accessible at http://127.0.0.1:8080

### Running with Docker Compose :cactus:

```bash
docker compose up
```

The server is accessible at http://127.0.0.1:8080

### Running in Kubernetes :dart:

#### With skaffold

Ensure you have a properly working and accessible Kubernetes cluster with a valid `~/.kube/config`. This project is using [Skaffold](https://skaffold.dev/) v2.2.0 to deploy to a local Kubernetes cluster, such as Minikube or KinD. You can dowload skaffold [here](https://skaffold.dev/docs/install/#standalone-binary). It is assumed that skaffold is installed.

To deploy to a local Kubernetes cluster, simply run skaffold run:

<details>

```
$ skaffold run
Generating tags...
 - clamav-api -> clamav-api:2023-07-09_01-24-29.9_CEST
Checking cache...
 - clamav-api: Not found. Building
Starting build...
Found [k3d-k3s-default] context, using local docker daemon.
Building [clamav-api]...
Target platforms: [linux/amd64]
#0 building with "default" instance using docker driver

#1 [internal] load build definition from Dockerfile
#1 transferring dockerfile: 256B done
#1 DONE 0.0s

#2 [internal] load .dockerignore
#2 transferring context: 122B done
#2 DONE 0.0s

#3 [internal] load metadata for docker.io/library/golang:1.20
#3 DONE 0.0s

#4 [builder 1/6] FROM docker.io/library/golang:1.20
#4 DONE 0.0s

#5 [internal] load build context
#5 transferring context: 10.14kB done
#5 DONE 0.0s

#6 [builder 2/6] WORKDIR /app
#6 CACHED

#7 [builder 3/6] COPY go.* ./
#7 CACHED

#8 [builder 4/6] RUN go mod download
#8 CACHED

#9 [builder 5/6] COPY . .
#9 DONE 0.1s

#10 [builder 6/6] RUN CGO_ENABLED=0 go build -ldflags '-d -w -s' -o main
#10 DONE 6.5s

#11 [stage-1 1/1] COPY --from=builder /app/main /
#11 CACHED

#12 exporting to image
#12 exporting layers done
#12 writing image sha256:c505b5f5ed48c79ec01a85bfb8827e036f211d17296e7b34aa94f9bb5e16a83d done
#12 naming to docker.io/library/clamav-api:2023-07-09_01-24-29.9_CEST done
#12 DONE 0.0s
Build [clamav-api] succeeded
Starting test...
Testing images...
Running custom test command: "go test ./..."
?   	github.com/lescactus/clamav-api-go	[no test files]
ok  	github.com/lescactus/clamav-api-go/internal/clamav	(cached)
?   	github.com/lescactus/clamav-api-go/internal/config	[no test files]
ok  	github.com/lescactus/clamav-api-go/internal/controllers	(cached)
ok  	github.com/lescactus/clamav-api-go/internal/logger	(cached)
Command finished successfully.
Tags used in deployment:
 - clamav-api -> clamav-api:c505b5f5ed48c79ec01a85bfb8827e036f211d17296e7b34aa94f9bb5e16a83d
Starting deploy...
Loading images into k3d cluster nodes...
 - clamav-api:c505b5f5ed48c79ec01a85bfb8827e036f211d17296e7b34aa94f9bb5e16a83d -> Found
Images loaded in 68.063635ms
 - configmap/clamav created
 - deployment.apps/clamav-api created
 - service/clamav-api created
 - serviceaccount/clamav-api created
Waiting for deployments to stabilize...
 - deployment/clamav-api: waiting for rollout to finish: 0 of 1 updated replicas are available...
 - deployment/clamav-api is ready.
Deployments stabilized in 31.094 seconds
You can also run [skaffold run --tail] to get the logs
```

</details>

The following happened:

* Skaffold will generate a docker tag based on the current timestamp.

* If the image doesn't exist locally, Skaffold will build it.

* Once the docker image is built, Skaffold will substitute the raw image defined in `deploy/k8s/deployment.yaml` with the image just built.

* Skaffold will apply the manifests in `deploy/k8s/`.

* Skaffold will wait for the `clamav-api` deployment to be ready.

For more informations about Skaffold and what it can do, visit the project [documentation](Without Skaffold).


#### Without Skaffold

To deploy to a Kubernetes cluster without Skaffold, simply build & push the docker image to an external registry. Then change the docker image name to include the registry in the `deploy/k8s/deployment.yaml` manifest.

## Specifications :ocean:

`GET /rest/v1/ping` will send the `PING` command to Clamd

`GET /rest/v1/version` will send the `VERSION` command to Clamd

`GET /rest/v1/stats` will send the `STATS` command to Clamd

`GET /rest/v1/versioncommands` will send the `VERSIONCOMMANDs` command to Clamd

`POST /rest/v1/reload` will send the `RELOAD` command to Clamd

`POST /rest/v1/shutdown` will send the `SHUTDOWN` command to Clamd

`POST /rest/v1/scan` will send the `INSTREAM` command to Clamd and stream the form for Clamd to scan. Note: this endpoint expects a `multipart/form-data`. See [Examples](https://github/com/lescactus/clamav-go-api#Examples) below.

## Examples

```
$ curl 127.0.0.1:8080/rest/v1/ping
{"ping":"PONG"}
```

```
$ curl 127.0.0.1:8080/rest/v1/version
{"clamav_version":"ClamAV 1.0.0/26734/Mon Nov 28 08:17:05 2022"}
```

```
$ curl 127.0.0.1:8080/rest/v1/stats
{"pools":1,"state":"VALID PRIMARY","threads":"live 1  idle 0 max 10 idle-timeout 30","queue":"0 items\n\tSTATS 0.000179 ","memstats":"heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 1260.177M pools_total 1260.222M"} 
```

```
$ curl 127.0.0.1:8080/rest/v1/versioncommands
{"clamav_version":"ClamAV 1.0.0/26734/Mon Nov 28 08:17:05 2022","commands":["SCAN","QUIT","RELOAD","PING","CONTSCAN","VERSIONCOMMANDS","VERSION","END","SHUTDOWN","MULTISCAN","FILDES","STATS","IDSESSION","INSTREAM","DETSTATSCLEAR","DETSTATS","ALLMATCHSCAN"]}
```

```
$ curl 127.0.0.1:8080/rest/v1/reload -XPOST
{"status":"RELOADING"}
```

```
$ curl 127.0.0.1:8080/rest/v1/shutdown -XPOST
{"status":"Shutting down"}
```

```
# Download the EICAR anti malware test file in /tmp/eicar.txt
$ wget https://secure.eicar.org/eicar.com.txt -O /tmp/eicar.txt

# Generate a 1M file with random content
$ dd if=/dev/urandom of=/tmp/test.txt bs=1M count=1

$ curl 127.0.0.1:8080/rest/v1/scan -F "file=@/tmp/test.txt" -v
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> POST /rest/v1/scan HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.81.0
> Accept: */*
> Content-Length: 1048762
> Content-Type: multipart/form-data; boundary=------------------------b6b6ef4a0ac767d7
> Expect: 100-continue
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 100 Continue
* We are completely uploaded and fine
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Content-Type: application/json
< X-Request-Id: cikv9kqrnmmc73e13940
< Date: Sat, 08 Jul 2023 23:44:19 GMT
< Content-Length: 74
< 
* Connection #0 to host 127.0.0.1 left intact
{"status":"noerror","msg":"stream: OK","signature":"","virus_found":false}

$ curl 127.0.0.1:8080/rest/v1/scan -F "file=@/tmp/eicar.txt" -v
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> POST /rest/v1/scan HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.81.0
> Accept: */*
> Content-Length: 255
> Content-Type: multipart/form-data; boundary=------------------------2c5ea3b07f1f700d
> 
* We are completely uploaded and fine
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Content-Type: application/json
< X-Request-Id: cikv9oirnmmc73e1394g
< Date: Sat, 08 Jul 2023 23:44:34 GMT
< Content-Length: 110
< 
* Connection #0 to host 127.0.0.1 left intact
{"status":"error","msg":"file contains potential virus","signature":"Win.Test.EICAR_HDB-1","virus_found":true}
```

## Development

air, skaffold unit tests


