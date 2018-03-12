package main

import (
	"./elevio"
	"./config"
	"./elevStateMap"
	"./fsm"
	"flag"
	"./network"
	"fmt"
)



//PACKET LOSS:

//sudo iptables -A INPUT -p udp -m statistic --mode random --probability 0.15 -j DROP
//



func main() {

//command line arguments for port and id
	var port string	
	var id string
	flag.StringVar(&id, "id", "", "id")
    flag.StringVar(&port, "port", "15657", "portnumber")
    
    flag.Parse()	
    if id == "" {
    	//dafault ID
		id = "0"
	}

	fmt.Println("id:", id)
    fmt.Println("port:", port)


	

//init fuctions 

	
	
	
//channels for communication between modules

	//hardware channels
	motorChan := make(chan config.MotorDirection)
	doorLampChan := make(chan bool)

    floorChan  := make(chan int)  
    buttonLampChan  := make(chan config.ButtonLamp, 20)
    statusChangesChan := make(chan config.ElevStateMap)
    orderChangesChan := make(chan config.ElevStateMap)
    newOrderChan := make(chan config.ButtonEvent)

    // We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
    peerUpdateCh := make(chan config.PeerUpdate)
    // This could be used to signal that we are somehow "unavailable".
    peerTxEnable := make(chan bool)

 
	messageTx := make(chan config.Message, 100)
	orderMsgRx := make(chan config.OrderMsg, 100)
	statusMsgRx := make(chan config.StatusMsg, 100)
	ackChan := make(chan config.AckMsg, 100)

	config.Init(id, port)
	elevio.InitDriver("localhost:" + port, config.NUM_FLOORS*)
	elevStateMap.InitElevStateMap(newOrderChan)
	
	init := true



    go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, orderChangesChan, newOrderChan, statusChangesChan)
    go elevio.Elevio(motorChan, doorLampChan, newOrderChan, floorChan, buttonLampChan)
	go network.Transmitter(16502, messageTx, ackChan)
	go network.Receiver(16502, orderMsgRx, statusMsgRx, ackChan, messageTx)
    go network.PeerTransmitter(15600, id, peerTxEnable)
	go network.PeerReceiver(15600, peerUpdateCh)


    
	
	for {
		select {
		case p := <-peerUpdateCh:

			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if init != true{
				fmt.Printf("EN NY PEEER JEG SENDER MINE ORDRE\n")
				network.SendOrders(messageTx, elevStateMap.GetLocalMap())	
			}

		case orderMsgFromNetwork := <- orderMsgRx:
			//Når vi mottar melding bør vi sjekke at hardware er oppdatert			
			elevStateMap.UpdateMapFromNetwork(orderMsgFromNetwork.ElevMap, newOrderChan, buttonLampChan)
			if init == true{
				elevio.InitOrders()
			}
			init = false

		case statusMsgFromNetwork := <- statusMsgRx:
			elevStateMap.UpdateElevStatusFromNetwork(statusMsgFromNetwork)

		case elevMap:= <-orderChangesChan:
			fmt.Printf("Sender ordremelding\n")
			elevStateMap.UpdateLocalMap(elevMap)
			//elevStateMap.PrintMap(elevMap)
			network.SendOrders(messageTx, elevMap)
			init = false

		case elevMap:= <-statusChangesChan:
			elevStateMap.UpdateLocalMap(elevMap)
			network.SendElevStatus(messageTx, elevMap)
			init = false
		}
	}
}