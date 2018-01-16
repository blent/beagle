![beagle](https://raw.githubusercontent.com/blent/beagle/master/assets/beagle-head-square-small.png)
# Beagle
> Beacons tracking system
![build](https://travis-ci.org/blent/beagle.svg?branch=master)

## Description
Beagle is a beacon tracking system that targets to run on small devices like Raspberry Pi.
It allows to track user-specific beacons and send notifications to dedicated RESTful services when they appear and/or disappear.

## Prerequisites

* [Go >= 1.6](https://golang.org/)
* [Glide package manager](https://github.com/Masterminds/glide)
* [GNU Make](https://www.gnu.org/software/make/)

### Linux

 * Kernel version 3.6 or above
 
### Windows
Not supported yet

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

### UI

There is a [UI Dashboard](https://github.com/blent/beagle-ui) for managing the system.    
In order to make Beagle serving static files from the project, just run it with the following options
```sh
beagle --http-static-dir PATH_TO_UI/dist/public
```

### Rest API

```beagle``` runs headlessly by default. In this case, unless HTTP is not disabled, all operations are made via REST API.

- ``GET /api/registry/peripherals`` - Returns a list of registered peripherals. Available query params: ``take:int``, ``skip:int``
- ``GET /api/registry/peripheral/:id`` - Returns a peripheral by a given id.
- ``POST /api/registry/peripheral`` - Creates a new peripheral.
- ``PUT /api/registry/peripheral/:id`` - Updates a peripheral by a given id.
- ``DELETE /api/registry/peripheral/:id`` - Deletes a single peripheral by a given id.
- ``DELETE /api/registry/peripherals`` - Deletes many peripherals by a given array of ids.

- ``GET    /api/registry/endpoints`` - Returns a list of registered endpoints. Available query params: ``take:int``, ``skip:int``
- ``GET    /api/registry/endpoint/:id`` - Returns an endpoint by a given id.
- ``POST   /api/registry/endpoint`` - Creates a new endpoint.
- ``PUT    /api/registry/endpoint`` - Updates an endpoint by a given id.
- ``DELETE /api/registry/endpoint/:id`` - Deletes a single endpoint by a given id.
- ``DELETE /api/registry/endpoints`` - Deletes many endpoints by a given array of ids.

- ``GET /api/monitoring/activity`` - Returns a list of active peripherals (registered and not registered). Available query params: ``take:int``, ``skip:int``

## Options

```sh
  -help
    	show this list
  -http
    	enables http server (default true)
  -http-api-route string
    	http server api route (default "/api")
  -http-port int
    	http server port number (default 8080)
  -http-static-dir string
    	http server static files directory
  -http-static-route string
    	http server static files route (default "/public")
  -name string
    	application name (default "beagle")
  -storage-connection string
    	storage connection string (default "/var/lib/beagle/database.db")
  -tracking-heartbeat int
    	peripheral heartbeat interval in seconds (default 5)
  -tracking-ttl int
    	peripheral ttl duration in seconds (default 5)
  -version
    	show version
```



