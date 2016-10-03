package main

import (
	"net"
	"time"
	"log"
	"net/rpc"
	"net/http"
	"strings"
)

func main() {
	ListenAndServe("9999")

}

func ListenAndServe(port string){
	chatRooms := make(map[string]Chatroom)
	clients := make(map[string]Client)
	cs := new(ChatServer)
	//chatRooms["name"]=Chatroom{"name",nil,nil}

	cs.chatRooms = chatRooms
	cs.clients = clients
	rpc.Register(cs)
	rpc.HandleHTTP()

	l,err := net.Listen("tcp", ":" + port)

	if err != nil {
		log.Panicf("Can't bind port to listen. %q", err)
	}

	go removeUnusedChatRooms(cs)
	http.Serve(l, nil)

}

func (c *ChatServer) ReceiveMessage(msg string, reply *string) error{
	log.Printf("Message Received: %v. CHATSERVER %v",msg,c)


	go broadcastMessage(msg, c)


	*reply = "hello"
	return nil
}

func (c *ChatServer) RegisterUser(msg string, reply *string) error{
	log.Printf("Setting User: %v. CHATSERVER %v",msg,c)

	if _, ok := c.clients[msg]; ok {
		*reply = "That username is already taken"
		return nil
	}

	client := new(Client)
	client.UserName = msg
	c.clients[client.UserName]=*client



	*reply = "Welcome " + client.UserName
	return nil
}

func broadcastMessage(msg string, c *ChatServer){
	user, message := parseMessage(msg)

	currentTime := time.Now()
	message = currentTime.Format("[Jan _2 15:04:05]")+" " + user + ": " + message

	chatRoomsToBroadcast := make([]string,0)

	for k := range c.chatRooms {
		flagContinue := true
		for i:= range c.chatRooms[k].clients{
			if c.chatRooms[k].clients[i]==user{
				chatRoomsToBroadcast = append(chatRoomsToBroadcast,k)
				flagContinue = false
				addMessageToChatroom(c,k,message)
				continue
			}
		}
		if !flagContinue{
			continue
		}
	}

	for k := range chatRoomsToBroadcast {
		//log.Printf("CHATROOM joined %v",c.chatRooms[k])
		chatRoomName := chatRoomsToBroadcast[k]
		for i:= range c.chatRooms[chatRoomName].clients{
			if c.chatRooms[chatRoomName].clients[i]!=user{
				messagesToDeliver := c.clients[c.chatRooms[chatRoomName].clients[i]].messagesToDeliver
				messagesToDeliver = append(messagesToDeliver,message)

				client := c.clients[c.chatRooms[chatRoomName].clients[i]]
				client.messagesToDeliver = messagesToDeliver

				c.clients[c.chatRooms[chatRoomName].clients[i]]= client
				//append(c.chatRooms[k].clients[i].messagesToDeliver,message)
			}
		}

	}
}

func (c *ChatServer) CreateChatRoom(msg string, reply *string) error {

	if _, ok := c.chatRooms[msg]; ok {
		*reply = "That chatRoom already exists"
		return nil
	}
	var clients []string
	var messages []Message

	chatRoom := Chatroom{msg,clients,messages}

	c.chatRooms[msg] = chatRoom

	*reply = "That chatRoom was created !"
	return nil

}

func (c *ChatServer) LeaveChatRoom(msg string, reply *string) error {

	user, chatroomName := parseMessage(msg)

	for k:= range  c.chatRooms{
		if(k==chatroomName){
			//Go througn all the clients for the chatroom
			for i:= range c.chatRooms[k].clients{
				//Is the user in this chatroom?
				if (c.chatRooms[k].clients[i] == user) {
					chatRoom := c.chatRooms[k]
					clients := chatRoom.clients
					clients = append(clients[:i],clients[i+1:]...)
					chatRoom.clients = clients
					c.chatRooms[k] = chatRoom
					*reply = "You left the ChatRoom"
					return nil
				}
			}
		}
	}
	*reply = "You are not in the chatroom or it doesn't exist"
	return nil
}

func (c *ChatServer) ListChatRoom(msg string, reply *string) error {

	for k:= range  c.chatRooms{
		*reply += "* " + k
	}
	return nil

}

func (c *ChatServer) JoinChatRoom(msg string, reply *string) error {


	user, chatroomName := parseMessage(msg)
	if _, ok := c.clients[user]; ok {
		if _, ok2 := c.chatRooms[chatroomName]; ok2 {
			chatRoom := c.chatRooms[chatroomName]
			chatRoomClients := chatRoom.clients

			for k := range chatRoomClients{
				if chatRoomClients[k] == user{
					*reply = "You are already joined to this chatroom"
					return nil
				}
			}
			chatRoomClients = append(chatRoomClients,user)
			chatRoom.clients = chatRoomClients
			c.chatRooms[chatroomName] = chatRoom

			*reply = "You joined the chatroom !"

			if len(c.chatRooms[chatroomName].messages)==1{
				*reply = *reply + "\n" + c.chatRooms[chatroomName].messages[0].message

			}else if len(c.chatRooms[chatroomName].messages)>1{
				var availableMessages string
				for _,msg:= range c.chatRooms[chatroomName].messages{
					availableMessages = availableMessages +  msg.message
				}

				*reply = *reply + "\n" + availableMessages
			}

			return nil
		}else{
			*reply = "That chatRoom doesn't exist :("
			return nil

		}
	}else{
		*reply = "That username doesn't exist!"
		return nil

	}

}


func (c *ChatServer) GetMessages(msg string, reply *string) error {


	user, _ := parseMessage(msg)

	if len(c.clients[user].messagesToDeliver)==0{
		return nil
	}else if len(c.clients[user].messagesToDeliver)==1{
		*reply = c.clients[user].messagesToDeliver[0]
		client := c.clients[user]

		client.messagesToDeliver = make([]string,0)

		c.clients[user] = client
	}else{
		var availableMessages string
		for _,msg:= range c.clients[user].messagesToDeliver{
			availableMessages = availableMessages +  msg
		}
		*reply = availableMessages

		client := c.clients[user]

		client.messagesToDeliver = make([]string,0)

		c.clients[user] = client
	}




	return nil

}


func removeUnusedChatRooms(cs *ChatServer){
	for{
		time.Sleep(1000* time.Millisecond)
		currentDate := time.Now
		var newChatRooms = make(map[string]Chatroom)
		var deletedFlag bool
		for k := range cs.chatRooms {
			var lastMessage Message
			//Go through all the messages for the chatroom
			for i:= range cs.chatRooms[k].messages{
				if lastMessage.message == ""{
					lastMessage=cs.chatRooms[k].messages[i]
				}else if lastMessage.date.Before(cs.chatRooms[k].messages[i].date){
					lastMessage=cs.chatRooms[k].messages[i]
				}else{
					lastMessage = lastMessage
				}


			}
			//if the chatroom has been used in the last 7 days or never been used, keep it
			if currentDate().AddDate(0,0,-7).Before(lastMessage.date) || lastMessage.message==""{
				log.Printf("Keeping: %v"+cs.chatRooms[k].name)
				newChatRooms[cs.chatRooms[k].name]=cs.chatRooms[k]
				deletedFlag = true
			}

		}
		if deletedFlag{
			cs.chatRooms = newChatRooms
		}


	}


}

func parseMessage(data string) (string, string){
	result := strings.SplitN(data, "/;",2)
	if len(result)<2{
		return result[0],""
	}
	return result[0],result[1]

}

func addMessageToChatroom(c *ChatServer, name, message string){
	chatRoom := c.chatRooms[name]
	newMessage := Message{message:message,date:time.Now()}
	chatRoom.messages = append(chatRoom.messages,newMessage)
	c.chatRooms[name] = chatRoom
}

type ChatServer struct {
	chatRooms	map[string]Chatroom
	clients		map[string]Client
}

type Client struct {
	UserName string
	messagesToDeliver []string
}

type Chatroom struct{
	name string
	clients []string
	messages []Message
}

type Message struct{
	message string
	date time.Time
}