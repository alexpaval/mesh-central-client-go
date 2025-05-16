package meshrouter

import (
	"crypto/tls"
	"io"
	"strings"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/soarinferret/mcc/internal/config"
)

type Device struct {
	Id     string
	Name   string
	OS     string
	IP     string
	Icon   int
	Conn   int
	Pwr    int
}

type Settings struct {
	ServerURL       string
	Username        string
	Password        string
	Token           string
	EmailToken      bool
	SMSToken        bool
	AuthCookie      string
	ServerID        string
	LoginKey        string
	LocalPort       int
	RemotePort      int
	RemoteTarget    string
	RemoteNodeID    string
	WebSocket       *websocket.Conn
	WebChannel      *websocket.Conn
	ACookie         string
	RCookie         string
	RenewCookieTimer *time.Timer
	ServerAuthClientNonce string
	MeshServerTlsHash string
	ServerHttpsHash string
	Devices		   []Device
	DeviceQueryState int
	debug		   bool
}

var settings Settings

func ApplySettings(remoteNodeId string, remotePort int, localPort int, remoteTarget string, debug bool){
	//settings.ServerURL = serverUrl
	//settings.Username = username
	//settings.Password = password
	settings.RemoteNodeID = remoteNodeId
	settings.RemotePort = remotePort
	settings.LocalPort = localPort
	settings.RemoteTarget = remoteTarget
	settings.debug = debug
}

func GetLocalPort() int {
	return settings.LocalPort
}

func StartSocket() {
	p := config.GetDefaultProfile()

	settings.Username = p.Username
	settings.Password = p.Password
	settings.ServerURL = "wss://" + p.Server + "/meshrelay.ashx"

	// Start by requesting a login token, this is needed because of 2FA and check that we have correct credentials from the start
	var options *url.URL
	var err error

	options, err = url.Parse(settings.ServerURL)
	if err != nil {
		fmt.Println("Unable to parse server URL.")
		os.Exit(1)
		return
	}

	xtoken := ""
	if settings.EmailToken {
		xtoken = "**email**"
	} else if settings.SMSToken {
		xtoken = "**sms**"
	} else if settings.Token != "" {
		xtoken = settings.Token
	}

	headers := http.Header{}
	if settings.ServerID == "" {
		if settings.AuthCookie != "" {
			options.RawQuery = fmt.Sprintf("auth=%s", settings.AuthCookie)
			if xtoken != "" {
				options.RawQuery += fmt.Sprintf("&token=%s", xtoken)
			}
		} else {
			auth := base64.StdEncoding.EncodeToString([]byte(settings.Username)) + "," +
				base64.StdEncoding.EncodeToString([]byte(settings.Password))
			if xtoken != "" {
				auth += "," + base64.StdEncoding.EncodeToString([]byte(xtoken))
			}
			headers.Add("x-meshauth", auth)
		}
	} else {
		headers.Add("x-meshauth", "*")
	}

	/*if settings.LoginKey != "" {
		options.RawQuery += fmt.Sprintf("&key=%s", settings.LoginKey)
	}*/



	// replace meshrelay.ashx with control.ashx
	urlStr := strings.Replace(settings.ServerURL, "meshrelay.ashx", "control.ashx", 1)

	//conn, _, err := websocket.DefaultDialer.Dial(urlStr, headers)
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	conn, _, err := dialer.Dial(urlStr, headers)
	if err != nil {
		fmt.Printf("Unable to connect to server: %v\n", err)
		os.Exit(1)
		return
	}

	if settings.debug {
		fmt.Println("Connected to server.")
	}

	settings.WebSocket = conn
	go onServerWebSocket(conn)

	// Keep the main function running
	//select {}
}

func onServerWebSocket(conn *websocket.Conn) {
	settings.WebChannel = conn

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// check if the error is a close message
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				if settings.debug {
					fmt.Println("Server closed connection")
				}
				return
			}
			fmt.Println("Server connection error:", err)
			return
		}

		var command map[string]interface{}
		if err := json.Unmarshal(message, &command); err != nil {
			fmt.Println("Error parsing command:", err)
			continue
		}

		switch command["action"] {
		case "close":
			handleCloseCommand(command)
		case "serverinfo":
			conn.WriteMessage(websocket.TextMessage, []byte(`{"action":"authcookie"}`))
		case "authcookie":
			handleAuthCookieCommand(command)
		case "serverAuth":
			handleServerAuthCommand(command)
		case "nodes":
			handleNodesCommand(command)
		}

	}
}

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

func handleCloseCommand(command map[string]interface{}) {
	if command["cause"] == "noauth" {
		switch command["msg"] {
		case "tokenrequired":
			if command["email2fasent"] == true {
				fmt.Println("Login token email sent.")
			} else if command["email2fa"] == true && command["sms2fa"] == true {
				fmt.Println("Login token required, use --token [token], or --emailtoken, --smstoken get a token.")
			} else if command["sms2fa"] == true {
				fmt.Println("Login token required, use --token [token], or --smstoken get a token.")
			} else if command["email2fa"] == true {
				fmt.Println("Login token required, use --token [token], or --emailtoken get a token.")
			} else {
				fmt.Println("Login token required, use --token [token].")
			}
		case "badtlscert":
			fmt.Println("Invalid TLS certificate detected.")
		case "badargs":
			fmt.Println("Invalid protocol arguments.")
		default:
			fmt.Println("Invalid username/password.")
		}
		os.Exit(1)
	} else {
		if settings.debug {
			fmt.Println("Server disconnected:", command["msg"])
		}
	}

}

func handleAuthCookieCommand(command map[string]interface{}) {
	if settings.ACookie == "" {
		settings.ACookie = command["cookie"].(string)
		settings.RCookie = command["rcookie"].(string)
		settings.RenewCookieTimer = time.AfterFunc(10*time.Minute, func() {
			settings.WebChannel.WriteMessage(websocket.TextMessage, []byte(`{"action":"authcookie"}`))
		})
		//startRouterEx()



	} else {
		settings.ACookie = command["cookie"].(string)
		settings.RCookie = command["rcookie"].(string)
	}
}

func handleServerAuthCommand(command map[string]interface{}) {
	// Switch to using HTTPS TLS certificate for authentication
	settings.ServerID = ""
	settings.ServerHttpsHash = settings.MeshServerTlsHash
	settings.MeshServerTlsHash = ""

	xtoken := ""
	if settings.EmailToken {
		xtoken = "**email**"
	} else if settings.SMSToken {
		xtoken = "**sms**"
	} else if settings.Token != "" {
		xtoken = settings.Token
	}

	auth := ""
	if settings.AuthCookie != "" {
		auth = fmt.Sprintf(`{"action":"userAuth","auth":"%s"`, settings.AuthCookie)
		if xtoken != "" {
			auth += fmt.Sprintf(`,"token":"%s"`, xtoken)
		}
		auth += "}"
	} else {
		auth = fmt.Sprintf(`{"action":"userAuth","username":"%s","password":"%s"`,
			base64.StdEncoding.EncodeToString([]byte(settings.Username)),
			base64.StdEncoding.EncodeToString([]byte(settings.Password)))
		if xtoken != "" {
			auth += fmt.Sprintf(`,"token":"%s"`, xtoken)
		}
		auth += "}"
	}

	settings.WebChannel.WriteMessage(websocket.TextMessage, []byte(auth))
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

func StopSocket() {
	// send close message
	settings.WebSocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "all done"))
}

func StartRouter(ready chan struct{}) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", settings.LocalPort))
	if err != nil {
		fmt.Printf("Unable to bind to local TCP port %d: %v\n", settings.LocalPort, err)
		os.Exit(1)
		return
	}
	settings.LocalPort = listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()

	close(ready)
	fmt.Printf("Redirecting local port %d to remote port %d.\n", listener.Addr().(*net.TCPAddr).Port, settings.RemotePort)
	fmt.Println("Press ctrl-c to exit.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go onTcpClientConnected(conn)
	}
}

func onTcpClientConnected(conn net.Conn) {
	if settings.debug {
		fmt.Println("Client connected")
	}
	defer conn.Close()

	conn.(*net.TCPConn).SetKeepAlive(true)
	conn.(*net.TCPConn).SetKeepAlivePeriod(30 * time.Second)

	options, err := url.Parse(fmt.Sprintf("%s?auth=%s&nodeid=%s&tcpport=%d",
		settings.ServerURL, settings.ACookie, settings.RemoteNodeID, settings.RemotePort))
	if err != nil {
		fmt.Println("Unable to parse server URL:", err)
		return
	}

	if settings.RemoteTarget != "" {
		options.RawQuery += fmt.Sprintf("&tcpaddr=%s", settings.RemoteTarget)
	}

	headers := http.Header{}
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	wsConn, _, err := dialer.Dial(options.String(), headers)
	if err != nil {
		fmt.Printf("Unable to connect to server: %v\n", err)
		return
	}

	go onWebSocket(wsConn, conn)

	select {}
}

func onWebSocket(wsConn *websocket.Conn, tcpConn net.Conn) {
	if settings.debug {
		fmt.Println("Websocket connected")
	}
	defer wsConn.Close()
	defer tcpConn.Close()

	// Channel to signal when either connection is closed
	done := make(chan struct{})
	var once sync.Once

	// Function to copy data from WebSocket to TCP
	go func() {
		defer once.Do(func() { close(done) })
		for {
			messageType, message, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					if settings.debug {
						fmt.Println("WebSocket closed normally")
					}
				} else {
					fmt.Println("WebSocket read error:", err)
				}
				return
			}
			if messageType == websocket.BinaryMessage && len(message) > 0 {
				_, err = tcpConn.Write(message)
				if err != nil {
					fmt.Println("TCP write error:", err)
					return
				}
			}
		}
	}()

	// Function to copy data from TCP to WebSocket
	go func() {
		defer once.Do(func() { close(done) })
		buf := make([]byte, 4096) // Buffer to read data in chunks
		for {
			n, err := tcpConn.Read(buf)
			if err != nil {
				if err == io.EOF {
					if settings.debug {
						fmt.Println("TCP connection closed by client")
					}
				} else {
					fmt.Println("TCP read error:", err)
				}
				return
			}
			if n > 0 {
				err = wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
				if err != nil {
					fmt.Println("WebSocket write error:", err)
					return
				}
			}
		}
	}()

	// Wait for either connection to be closed
	<-done
}
