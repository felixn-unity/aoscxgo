package aoscxgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

type LagInterface struct {

	// Connection properties.
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	AdminState       string                 `json:"admin"`
	VlanMode         string                 `json:"vlan_mode"`
	VlanIds          []interface{}          `json:"vlan_ids"`
	VlanTag          int                    `json:"vlan_tag"`
	TrunkAllowedAll  bool                   `json:"trunk_allowed_all"`
	NativeVlanTag    bool                   `json:"native_vlan_tag"`
	LacpMode         string                 `json:"lacp_mode"`
	InterfaceDetails map[string]interface{} `json:"details"`
	materialized     bool                   `json:"materialized"`
	uri              string                 `json:"uri"`
}

// checkValues validates LAG interface configuration
func (l *LagInterface) checkValues() error {
	// Validate LAG name format
	if matched, _ := regexp.MatchString("^lag\\d+$", l.Name); !matched {
		return &RequestError{
			StatusCode: "Invalid Required Value: Name - must be in format 'lagXX' (e.g., lag60)",
			Err:        errors.New("validation error"),
		}
	}

	// Validate admin state
	if l.AdminState != "up" && l.AdminState != "down" {
		return &RequestError{
			StatusCode: "Invalid Required Value: AdminState - valid options are 'up' or 'down' received: " + l.AdminState,
			Err:        errors.New("validation error"),
		}
	}

	// Validate LACP mode if provided
	if l.LacpMode != "" && l.LacpMode != "active" && l.LacpMode != "passive" {
		return &RequestError{
			StatusCode: "Invalid Required Value: LacpMode - valid options are 'active' or 'passive' received: " + l.LacpMode,
			Err:        errors.New("validation error"),
		}
	}

	return nil
}

// ensureVlanExists checks if VLAN exists, creates it if not
func (l *LagInterface) ensureVlanExists(c *Client, vlanId int) (*Vlan, error) {
	vlan := &Vlan{VlanId: vlanId}
	err := vlan.Get(c)

	if err != nil && !vlan.materialized {
		err = vlan.Create(c)
		if err != nil && !vlan.materialized {
			return nil, &RequestError{
				StatusCode: "VLAN " + strconv.Itoa(vlanId) + " not found and unable to create",
				Err:        errors.New("vlan dependency error"),
			}
		}
	}
	return vlan, nil
}

// buildVlanConfig constructs VLAN configuration for the interface
func (l *LagInterface) buildVlanConfig(c *Client) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	if l.VlanMode == "" || l.VlanMode == "access" {
		// Access mode configuration
		vlanId := l.VlanTag
		if vlanId == 0 {
			vlanId = 1
		}

		vlan, err := l.ensureVlanExists(c, vlanId)
		if err != nil {
			return nil, err
		}

		config["vlan_tag"] = map[string]interface{}{strconv.Itoa(vlanId): vlan.GetURI()}
		config["vlan_mode"] = "access"

	} else if l.VlanMode == "trunk" || l.VlanMode == "native-untagged" || l.VlanMode == "native-tagged" {
		// Trunk mode configuration
		config["vlan_mode"] = l.VlanMode

		// Configure native VLAN
		if l.VlanTag == 0 || l.VlanTag == 1 {
			config["vlan_tag"] = nil
		} else {
			vlan, err := l.ensureVlanExists(c, l.VlanTag)
			if err != nil {
				return nil, err
			}
			config["vlan_tag"] = map[string]interface{}{strconv.Itoa(l.VlanTag): vlan.GetURI()}
		}

		// Configure trunk VLANs
		vlanTrunks := make(map[string]interface{})
		if len(l.VlanIds) > 0 {
			for _, item := range l.VlanIds {
				vlanId := item.(int)
				vlan, err := l.ensureVlanExists(c, vlanId)
				if err == nil {
					vlanTrunks[strconv.Itoa(vlanId)] = vlan.GetURI()
				}
			}
		}
		config["vlan_trunks"] = vlanTrunks

	} else {
		return nil, &RequestError{
			StatusCode: "Invalid VlanMode: " + l.VlanMode + " - valid options are 'access', 'trunk', 'native-untagged', or 'native-tagged'",
			Err:        errors.New("validation error"),
		}
	}

	return config, nil
}

// Create performs POST to create LAG Interface configuration
func (l *LagInterface) Create(c *Client) error {
	if err := l.checkValues(); err != nil {
		return err
	}

	baseURI := "system/interfaces"
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + baseURI
	l.uri = "/rest/" + c.Version + "/" + baseURI + "/" + l.Name

	// Build base configuration
	postMap := map[string]interface{}{
		"name":        l.Name,
		"type":        "lag",
		"description": l.Description,
		"admin":       l.AdminState,
	}

	// Add LACP configuration
	if l.LacpMode != "" {
		postMap["lacp"] = l.LacpMode
	}

	// Add VLAN configuration
	if l.VlanMode != "" {
		vlanConfig, err := l.buildVlanConfig(c)
		if err != nil {
			return err
		}
		for key, value := range vlanConfig {
			postMap[key] = value
		}
	}

	// Execute POST request
	postBody, _ := json.Marshal(postMap)
	jsonBody := bytes.NewBuffer(postBody)

	res := post(c, url, jsonBody)
	if res.Status != "201 Created" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("create error"),
		}
	}

	l.materialized = true
	return nil
}

// Update performs PATCH or PUT to update LAG Interface configuration
func (l *LagInterface) Update(c *Client, usePut bool) error {
	if err := l.checkValues(); err != nil {
		return err
	}

	if l.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface Name",
			Err:        errors.New("update error"),
		}
	}

	baseURI := "system/interfaces"
	intStr := url.PathEscape(l.Name)
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + baseURI + "/" + intStr

	updateMap := make(map[string]interface{})

	// For PUT, get existing configuration
	if usePut {
		tmpLag := LagInterface{Name: l.Name}
		if err := tmpLag.Get(c); err != nil {
			return err
		}
		for key, value := range tmpLag.InterfaceDetails {
			updateMap[key] = value
		}
	}

	// Set basic properties
	updateMap["description"] = l.Description
	updateMap["admin"] = l.AdminState

	// Add LACP configuration
	if l.LacpMode != "" {
		updateMap["lacp"] = l.LacpMode
	}

	// Add VLAN configuration
	if l.VlanMode != "" {
		vlanConfig, err := l.buildVlanConfig(c)
		if err != nil {
			return err
		}
		for key, value := range vlanConfig {
			updateMap[key] = value
		}
	}

	// Execute request
	updateBody, _ := json.Marshal(updateMap)
	jsonBody := bytes.NewBuffer(updateBody)

	var res *http.Response
	if usePut {
		res = put(c, url, jsonBody)
		if res.Status != "200 OK" {
			return &RequestError{
				StatusCode: "PUT failed: " + res.Status,
				Err:        errors.New("update error"),
			}
		}
	} else {
		res = patch(c, url, jsonBody)
		if res.Status != "204 No Content" {
			return &RequestError{
				StatusCode: "PATCH failed: " + res.Status,
				Err:        errors.New("update error"),
			}
		}
	}

	l.materialized = true
	return nil
}

// Delete removes LAG Interface configuration
func (l *LagInterface) Delete(c *Client) error {
	if l.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface Name",
			Err:        errors.New("delete error"),
		}
	}

	baseURI := "system/interfaces"
	intStr := url.PathEscape(l.Name)
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + baseURI + "/" + intStr

	res := delete(c, url)

	if res.Status != "204 No Content" && res.Status != "200 OK" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("delete error"),
		}
	}

	l.materialized = false
	return nil
}

// Get retrieves LAG Interface configuration
func (l *LagInterface) Get(c *Client) error {
	if l.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface Name",
			Err:        errors.New("get error"),
		}
	}

	baseURI := "system/interfaces"
	intStr := url.PathEscape(l.Name)
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + baseURI + "/" + intStr + "?selector=writable"

	res, body := get(c, url)

	if res.Status != "200 OK" {
		l.materialized = false
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("get error"),
		}
	}

	// Initialize details map
	if l.InterfaceDetails == nil {
		l.InterfaceDetails = make(map[string]interface{})
	}

	// Parse response
	for key, value := range body {
		l.InterfaceDetails[key] = value

		switch key {
		case "description":
			if value != nil {
				l.Description = value.(string)
			}
		case "admin":
			if value != nil {
				l.AdminState = value.(string)
			}
		case "lacp":
			if value != nil {
				l.LacpMode = value.(string)
			}
		case "vlan_mode":
			if value != nil {
				l.VlanMode = value.(string)
				l.NativeVlanTag = (l.VlanMode == "native-tagged")
			}
		case "vlan_tag":
			if value != nil {
				for vlanStr := range value.(map[string]interface{}) {
					if vlanInt, err := strconv.Atoi(vlanStr); err == nil {
						l.VlanTag = vlanInt
					}
				}
			}
		case "vlan_trunks":
			vlanMap := value.(map[string]interface{})
			if len(vlanMap) > 0 {
				var vlanIds []interface{}
				for vlanStr := range vlanMap {
					if vlanInt, err := strconv.Atoi(vlanStr); err == nil {
						vlanIds = append(vlanIds, vlanInt)
					}
				}
				l.VlanIds = vlanIds
				l.TrunkAllowedAll = false
			} else {
				l.TrunkAllowedAll = true
				l.VlanIds = []interface{}{}
			}
		}
	}

	l.materialized = true
	return nil
}

// GetStatus returns True if LAG Interface exists on Client object or False if not
func (l *LagInterface) GetStatus() bool {
	return l.materialized
}

// GetURI returns URI of LAG Interface
func (l *LagInterface) GetURI() string {
	return l.uri
}
