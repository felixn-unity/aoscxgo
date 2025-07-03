package main

import (
	"log"
	"os"
	"strconv"

	"github.com/felixn-unity/aoscxgo"
)

func main() {
	// Get configuration from environment variables
	hostname := os.Getenv("AOSCX_HOSTNAME")
	if hostname == "" {
		log.Fatal("AOSCX_HOSTNAME environment variable is required")
	}

	username := os.Getenv("AOSCX_USERNAME")
	if username == "" {
		username = "admin" // Default username
	}

	password := os.Getenv("AOSCX_PASSWORD")
	if password == "" {
		log.Fatal("AOSCX_PASSWORD environment variable is required")
	}

	// Optional: Certificate verification (defaults to false for development)
	verifyCert := false
	if certStr := os.Getenv("AOSCX_VERIFY_CERT"); certStr != "" {
		if parsed, err := strconv.ParseBool(certStr); err == nil {
			verifyCert = parsed
		}
	}

	// Optional: API version (defaults to v10.09)
	version := os.Getenv("AOSCX_VERSION")
	if version == "" {
		version = "v10.09"
	}

	sw, err := aoscxgo.Connect(
		&aoscxgo.Client{
			Hostname:          hostname,
			Username:          username,
			Password:          password,
			Version:           version,
			VerifyCertificate: verifyCert,
		},
	)

	if err != nil || sw == nil {
		log.Printf("Failed to login to switch: %s", err)
		return
	}

	log.Printf("Successfully connected to switch %s", hostname)

	lagdel := aoscxgo.LagInterface{
		Name: "lag60",
	}

	if 1 == 2 {
		err = lagdel.Delete(sw)
		if err != nil {
			log.Printf("Error in creating LAG 60: %s", err)
			return
		}
		log.Printf("LAG Deleted Success")

		lagdel = aoscxgo.LagInterface{
			Name: "lag70",
		}

		err = lagdel.Delete(sw)
		if err != nil {
			log.Printf("Error in creating LAG 60: %s", err)
			return
		}
		log.Printf("LAG Deleted Success")
	}

	if 1 == 2 {
		lagnew := aoscxgo.LagInterface{
			Name:            "lag60",
			Description:     "uplink VLAN",
			AdminState:      "up",
			VlanMode:        "native-untagged",
			VlanTag:         200,
			VlanIds:         []interface{}{100, 200},
			TrunkAllowedAll: true,
			NativeVlanTag:   true,
			LacpMode:        "active",
		}

		err = lagnew.Create(sw)
		if err != nil {
			log.Printf("Error in creating LAG 60: %s", err)
			return
		}
		log.Printf("LAG Create Success")
		lagnew = aoscxgo.LagInterface{
			Name:          "lag70",
			Description:   "uplink VLAN",
			AdminState:    "up",
			VlanMode:      "access",
			VlanTag:       200,
			NativeVlanTag: true,
			LacpMode:      "active",
		}

		err = lagnew.Create(sw)
		if err != nil {
			log.Printf("Error in creating LAG 60: %s", err)
			return
		}
		log.Printf("LAG Create Success")
	}

	if 1 == 2 {

		vlandel := aoscxgo.Vlan{
			VlanId: 600,
		}

		err = vlandel.Delete(sw)

		if err != nil {
			log.Printf("Error in creating VLAN 100: %s", err)
			return
		}

		log.Printf("VLAN Delete Success")
	}

	if 1 == 1 {

		vlan := aoscxgo.Vlan{
			VlanId:      600,
			Name:        "uplink VLAN",
			Description: "uplink VLAN",
			AdminState:  "up",
		}

		// if the vlan exists use
		// err = vlan100.Update(sw)
		err = vlan.Create(sw)

		if err != nil {
			log.Printf("Error in creating VLAN 100: %s", err)
			return
		}

		log.Printf("VLAN Create Success")
	}

	sw.Logout()

}
