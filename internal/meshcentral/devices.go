package meshcentral

import (
	"fmt"
	"time"
	"github.com/gorilla/websocket"
)


func handleNodesCommand(command map[string]interface{}) {
	if settings.debug {
		fmt.Println("Received nodes command")
	}
	var devices []Device
	nodeGroups := command["nodes"].(map[string]interface{})
	for _, nodeGroup := range nodeGroups {
		nodes := nodeGroup.([]interface{})
		for _, node := range nodes {
			nodeMap := node.(map[string]interface{})
			// check to see if items are blank, if so, set to ""
			if nodeMap["name"] == nil {
				nodeMap["name"] = ""
			}
			if nodeMap["osdesc"] == nil {
				nodeMap["osdesc"] = ""
			}
			if nodeMap["ip"] == nil {
				nodeMap["ip"] = ""
			}

			if nodeMap["pwr"] == nil {
				nodeMap["pwr"] = 0.0
			}
			if nodeMap["conn"] == nil {
				nodeMap["conn"] = 0.0
			}
			device := Device{
				Id:     nodeMap["_id"].(string),
				Name:   nodeMap["rname"].(string),
				OS:     nodeMap["osdesc"].(string),
				IP:     nodeMap["ip"].(string),
				Icon:   int(nodeMap["icon"].(float64)),
				//Conn:   0,
				Conn:   int(nodeMap["conn"].(float64)),
				//Pwr:	0,
				Pwr:    int(nodeMap["pwr"].(float64)),
			}
			devices = append(devices, device)
		}
	}

	settings.Devices = devices
	settings.DeviceQueryState = 0
}

// hacky until I get better at golang
func GetDevices() []Device {
	settings.DeviceQueryState = 1
	settings.WebSocket.WriteMessage(websocket.TextMessage, []byte(`{"action":"nodes"}`))

	// wait for devices to be populated
	for settings.DeviceQueryState == 1 {
		time.Sleep(250 * time.Millisecond)
	}

	return settings.Devices
}
