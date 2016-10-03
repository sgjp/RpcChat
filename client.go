package main

import (
	"net/rpc"
	"log"
	"fmt"
	"bufio"
	"os"
	"strings"
	"time"
)

var userName string

func main() {
	conn, err := rpc.DialHTTP("tcp", ":9999")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	// Synchronous call
	setUserName(conn)

	go getMessages(conn)

	showMenu()


	for true {
		InputHandler(conn)
	}

}



func InputHandler(conn *rpc.Client){
	reader := bufio.NewReader(os.Stdin)

	for true {
		m, _ := reader.ReadString('\n')
		option, args := parseInput(m)
		if(option==""){
			fmt.Printf("Please select an option. Remember: 0 shows the menu")
		}else{
			option = strings.Replace(option,"\n","",-1)
			switch option {

			//Show the menu
			case "0":
				showMenu()


			//Create chatroom
			case "1":
				if(args==""){
					fmt.Printf("Not Args found. Example: '1 NewChatRoom'")
				}else{
					var reply string
					err := conn.Call("ChatServer.CreateChatRoom", args, &reply)
					fmt.Printf("%v\n", reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
					//message := Encode(args)
					//conn.Write([]byte("C/;"+args))

				}

			//List chatrooms
			case "2":
				var reply string
				err := conn.Call("ChatServer.ListChatRoom", args, &reply)
				if err != nil {
					log.Fatal("Error:", err)
				}
				fmt.Printf("%v", reply)



			//Join Existing chatroom
			case "3":
				if(args==""){
					fmt.Println("Not Args found. Example: '3 ExistingChatRoom'")
				}else{
					var reply string
					err := conn.Call("ChatServer.JoinChatRoom", userName+"/;"+args, &reply)
					fmt.Printf("%v\n", reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
				}

			//Send message
			case "4":
				if(args==""){
					fmt.Println("Not Args found. Example: '4 hello everyone!'")
				}else{
					//message := Encode(args)
					var reply string
					err := conn.Call("ChatServer.ReceiveMessage", userName+"/;"+args, &reply)
					if err != nil {
						log.Fatal("Error:", err)
					}

					//conn.Write([]byte("M/;"+args))
				}

			//Leave chatroom
			case "5":
				if(args==""){
					fmt.Println("Not Args found. Example: '5 ExistingChatRoom'")
				}else{
					var reply string
					err := conn.Call("ChatServer.LeaveChatRoom", userName+"/;"+args, &reply)
					fmt.Printf("%v\n", reply)
					if err != nil {
						log.Fatal("Error:", err)
					}
				}

			default:
			//conn.Write([]byte("M/;"+args))

			}

		}
	}
}
func showMenu(){
	fmt.Println("")
	fmt.Println("--PLEASE SELECT THE DESIRED OPTION:\n")
	fmt.Println("  1. Create a chatroom.   Args: Name")
	fmt.Println("  2. List chatrooms.")
	fmt.Println("  3. Join existing chatroom.   Args: Name")
	fmt.Println("  4. Send Message to all joined chatrooms  Args: Message")
	fmt.Println("  5. Quit chatroom.    Args: Name")
	fmt.Println("  0. Show Menu")
	fmt.Println("")
	fmt.Println("  Example:  '3 chatroom2'")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")

}

func setUserName(conn *rpc.Client){
	fmt.Println("Please set your username:")
	reader := bufio.NewReader(os.Stdin)
	userName, _ = reader.ReadString('\n')

	userName = strings.Replace(userName,"\n","",-1)

	var reply string
	err := conn.Call("ChatServer.RegisterUser", userName, &reply)
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Printf("%v\n", reply)

}

func getMessages(conn *rpc.Client){
	for{
		var reply string
		err := conn.Call("ChatServer.GetMessages", userName, &reply)
		if err != nil {
			log.Printf("Error:", err)
		}
		if reply != "" {
			fmt.Printf("%v", reply)
		}

		time.Sleep(300*time.Millisecond)

	}

}

func parseInput(m string)(string, string){
	splitted := strings.SplitN(m," ",2)
	if(len(splitted)>1){
		return splitted[0],splitted[1]
	}
	if(len(splitted)==1){
		return splitted[0],""
	}
	return "",""
}