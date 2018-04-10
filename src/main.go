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
	motorChan := make(chan config.MotorDirection, 100)
	doorLampChan := make(chan bool, 100)

    floorChan  := make(chan int, 100)  
    buttonLampChan  := make(chan config.ButtonLamp, 100)
    statusChangesChan := make(chan config.ElevStateMap, 100)
    mapChangesChan := make(chan config.ElevStateMap, 100)
    newOrderChan := make(chan config.ButtonEvent, 100)

    // We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
    peerUpdateCh := make(chan config.PeerUpdate, 100)
    // This could be used to signal that we are somehow "unavailable".
    peerTxEnable := make(chan bool, 100)

 
	messageTx := make(chan config.Message, 10000)
	orderMsgRx := make(chan config.OrderMsg, 10000)
	statusMsgRx := make(chan config.StatusMsg, 10000)
	ackChan := make(chan config.AckMsg, 10000)
	
	config.Init(id, port)
	elevio.InitDriver("localhost:" + port, config.NUM_FLOORS)
	elevStateMap.InitElevStateMap()
	elevio.InitOrders()

	fmt.Printf("Init success\n")
	
	init := true



  
    go elevio.Elevio(motorChan, doorLampChan, newOrderChan, floorChan, buttonLampChan)
    go elevio.OrderLights(newOrderChan, buttonLampChan)
	go network.Transmitter(16502, messageTx, ackChan)
	go network.Receiver(16502, orderMsgRx, statusMsgRx, ackChan, messageTx)
    go network.PeerTransmitter(15600, id, peerTxEnable)
	go network.PeerReceiver(15600, peerUpdateCh)
	go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, mapChangesChan, newOrderChan, statusChangesChan)

	fmt.Printf("go all functions\n")
    
	
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
			orderUpdates := elevStateMap.UpdateMapFromNetwork(orderMsgFromNetwork.ElevMap, newOrderChan, buttonLampChan)
			if init == true{
				elevio.InitOrders()
			}
			init = false

			if orderUpdates {
				network.SendOrders(messageTx, elevStateMap.GetLocalMap())
			}

		case statusMsgFromNetwork := <- statusMsgRx:
			elevStateMap.UpdateElevStatusFromNetwork(statusMsgFromNetwork)
			

		case elevMap:= <-mapChangesChan:
			//fmt.Printf("Sender ordremelding\n")
			localOrderUpdates := elevStateMap.UpdateLocalMap(elevMap)
			//elevStateMap.PrintMap(elevMap)
			if localOrderUpdates {
				network.SendOrders(messageTx, elevMap)
			}
			network.SendElevStatus(messageTx, elevMap)

		
			init = false

/*		case elevMap:= <-statusChangesChan:
			elevStateMap.UpdateLocalMap(elevMap)
			network.SendElevStatus(messageTx, elevMap)
			init = false*/
		}
	}
}