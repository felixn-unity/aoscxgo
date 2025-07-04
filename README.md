aoscxgo
========================

aoscxgo is a golang package that allows users to connect to and configure AOS-CX switches using REST API. The minimum supported firmware version is 10.09.

## Features

- **Automatic API Version Detection**: Automatically detects and uses the latest API version supported by the switch
- **Environment Variable Configuration**: Secure configuration using environment variables
- **Comprehensive Error Handling**: Robust error handling with detailed error messages
- **Full CRUD Operations**: Create, Read, Update, Delete operations for all supported resources

Configuration
=============

## Environment Variables

The recommended way to configure the switch connection is using environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AOSCX_HOSTNAME` | Yes | - | Switch hostname or IP address |
| `AOSCX_PASSWORD` | Yes | - | Switch password |
| `AOSCX_USERNAME` | No | `admin` | Switch username |
| `AOSCX_VERSION` | No | `auto-detect` | API version (will auto-detect if not specified) |
| `AOSCX_VERIFY_CERT` | No | `false` | Enable SSL certificate verification |

### API Version Detection

The library automatically detects the latest API version supported by the switch by querying `https://<hostname>/rest`. This ensures optimal compatibility and access to the latest features. If auto-detection fails, it falls back to:

1. User-specified `AOSCX_VERSION` environment variable
2. Default version `v10.09`

### Setup Environment Variables

1. Copy the example environment file:
   ```bash
   cp env.example .env
   ```

2. Edit `.env` with your switch details:
   ```bash
   AOSCX_HOSTNAME=192.168.1.100
   AOSCX_PASSWORD=your_password_here
   AOSCX_USERNAME=admin
   # AOSCX_VERSION=v10.09  # Optional - will auto-detect if not specified
   AOSCX_VERIFY_CERT=false
   ```

3. Source the environment file:
   ```bash
   source .env
   ```

Using aoscxgo
===========

## Basic Connection

The application will automatically read configuration from environment variables and detect the best API version:

```go
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/felixn-unity/aoscxgo"
)

func main() {
	// Configuration is read from environment variables
	hostname := os.Getenv("AOSCX_HOSTNAME")
	if hostname == "" {
		log.Fatal("AOSCX_HOSTNAME environment variable is required")
	}

	username := os.Getenv("AOSCX_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("AOSCX_PASSWORD")
	if password == "" {
		log.Fatal("AOSCX_PASSWORD environment variable is required")
	}

	// The client will automatically detect the latest API version
	sw, err := aoscxgo.Connect(
		&aoscxgo.Client{
			Hostname:          hostname,
			Username:          username,
			Password:          password,
			VerifyCertificate: false,
		},
	)

	if err != nil || sw == nil {
		log.Printf("Failed to login to switch: %s", err)
		return
	}
	log.Printf("Successfully connected to switch %s using API version %s", hostname, sw.Version)
}
```

## Manual Configuration (Alternative)

You can also configure the client manually and specify a particular API version:

```go
sw, err := aoscxgo.Connect(
	&aoscxgo.Client{
		Hostname:          "10.0.0.1",
		Username:          "admin",
		Password:          "admin",
		Version:           "v10.10", // Force specific version
		VerifyCertificate: false,
	},
)
```

## VLAN Management Example

This will login to the switch and create a cookie to use for authentication in further calls. This cookie is stored within the aoscxgo.Client object that will be passed into configuration modules like so:

```go
	vlan100 := aoscxgo.Vlan{
		VlanId:      100,
		Name:        "uplink VLAN",
		Description: "uplink VLAN",
		AdminState:  "up",
	}

	// if the vlan exists use
	// err = vlan100.Update(sw)
	err = vlan100.Create(sw)

	if err != nil {
		log.Printf("Error in creating VLAN 100: %s", err)
		return
	}

	log.Printf("VLAN Create Success")
```

API Methods
===========

Each API resource will have the following functions (exceptions may vary):

  * `Create()`
  * `Update()`
  * `Get()`
  * `GetStatus()`
  * `Delete()`

## Running the Example

1. Set up your environment variables:
   ```bash
   export AOSCX_HOSTNAME="192.168.1.100"
   export AOSCX_PASSWORD="your_password"
   export AOSCX_USERNAME="admin"
   ```

2. Run the example:
   ```bash
   go run ./cmd/main.go
   ```

The application will automatically:
- Detect the latest API version supported by your switch
- Connect using the optimal API version
- Display the detected version in the log output
