![beagle](https://raw.githubusercontent.com/blent/beagle/master/assets/beagle-head-square-small.png)
# Beagle
> Beacons tracking system

## Description
Beagle is a beacon tracking system that targets to run on small devices like Raspberry Pi.
It allows to track user-specific beacons and send notifications to dedicated RESTful services when they appear and/or disappear.

## Prerequisites

### Linux

 * [Glide package manager](https://github.com/Masterminds/glide)
 * Kernel version 3.6 or above

### macOS

 * [Glide package manager](https://github.com/Masterminds/glide) 

## Installation

```sh
git clone https://github.com/blent/beagle
cd beagle
make build
```

### Cross-compile and deploy to a target device

Build and run Beagle on a ARMv5 target device.
```sh
GOARCH=arm GOARM=5 GOOS=linux go build -v -o ./bin/beagle ./src/main.go
```

## Start

Since Beagle programs administer network devices, they must either be run as root, or be granted appropriate capabilities:

```sh
sudo beagle
```

### Options

```sh
  -http
    	enables http server (default true)
  -http-api-route string
    	http server api route (default "/api")
  -http-port int
    	htpp server port number (default 8080)
  -http-static-dir string
    	http server static files directory
  -http-static-route string
    	http server static files route
  -name string
    	application name (default "beagle")
  -storage-connection string
    	storage connection string (default "/var/lib/beagle/database.db")
  -tracking-heartbeat int
    	peripheral heartbeat interval in seconds (default 5)
  -tracking-ttl int
    	peripheral ttl duration in seconds (default 5)
```



