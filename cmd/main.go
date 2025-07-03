package main

import (
	"log"

	"github.com/felixn-unity/aoscxgo"
)

func main() {
	sw, err := aoscxgo.Connect(
		&aoscxgo.Client{
			Hostname:          "yyyy",
			Username:          "admin",
			Password:          "xxxx",
			VerifyCertificate: false,
		},
	)

	if (sw.Cookie == nil) || (err != nil) {
		log.Printf("Failed to login to switch: %s", err)
		return
	}
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
