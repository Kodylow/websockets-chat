package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)
var clients = make(map[WsConnection]string)

// boilerplate for Jet
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Home displays the homepage
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.html", nil)
	if err != nil {
		log.Println(err)
	}
}

// WsConnection allows access to the gmux ws connector
type WsConnection struct {
	*websocket.Conn
}

// WsJSONResponse defines the response sent back from websocket
type WsJSONResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// WsPayload defines the payload
type WsPayload struct {
	Action   string       `json:"action"`
	Username string       `json:"username"`
	Message  string       `json:"message"`
	Conn     WsConnection `json:"-"`
}

// WsEndpoint creates the upgraded websocket connection off of the http req res
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	var response WsJSONResponse
	response.Message = `<em><small>Connected to server</small></em>`

	//creates that websocket connection
	conn := WsConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	//start the goroutine listening for websocket payloads
	go ListenForWs(&conn)
}

// ListenForWs listens for a websocket payload, which it shoots off into a channel
func ListenForWs(conn *WsConnection) {
	// restarts if the listener dies
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	//creates a listening loop, everytime it gets a req with a payload it passes it into the channel
	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			//pass into channel
			wsChan <- payload
		}
	}
}

// ListenToWsChannel , goroutine, listens for payloads passed into the channel
func ListenToWsChannel() {
	var response WsJSONResponse

	for {
		//get a payload from the channel
		event := <-wsChan

		switch event.Action {
		case "username":
			//get a list of all users and send it back via broadcast
			clients[event.Conn] = event.Username
			users := getUserList()

			//fill out the response
			response.Action = "list_users"
			response.ConnectedUsers = users

			//broadcast to everyone
			broadcast(response)

		// client left, remove from users and send back updated userlist
		case "left":
			response.Action = "list_users"
			delete(clients, event.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcast(response)

		case "broadcast":
			response.Action = "broadcast"
			//format as a from: message sorta look
			response.Message = fmt.Sprintf("<strong>%s</strong>, %s", event.Username, event.Message)
			broadcast(response)
		}
	}
}

// getUserList grabs users off clients map and returns as an array
func getUserList() []string {
	var userList []string
	for _, x := range clients {
		//only add if not a lurker
		if x != "" {
			userList = append(userList, x)
		}
	}

	sort.Strings(userList)
	return userList
}

// Broadcast forwards the payload as json writing to all the other clients when someone posts in the chatroom
func broadcast(response WsJSONResponse) {
	//for every client that you know about, send them this response
	for client := range clients {

		//writes the Json to the client
		err := client.WriteJSON(response)
		//err means we've lost the client
		if err != nil {
			log.Println("websocket err, lost client")
			//close the websocket
			_ = client.Close()
			//remove the client from map of current clients
			delete(clients, client)
		}
	}
}

// renderPage renders the page using jet templating
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
