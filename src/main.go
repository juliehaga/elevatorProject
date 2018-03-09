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
    buttonLampChan  := make(chan config.ButtonLamp)
    statusChangesChan := make(chan config.ElevStateMap)
    orderChangesChan := make(chan config.ElevStateMap)
    newOrderChan := make(chan config.ButtonEvent)

    // We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
    peerUpdateCh := make(chan config.PeerUpdate)
    // This could be used to signal that we are somehow "unavailable".
    peerTxEnable := make(chan bool)


 
	messageTx := make(chan config.Message)
	orderMsgRx := make(chan config.OrderMsg)
	statusMsgRx := make(chan config.StatusMsg)

	config.Init(id, port)
	elevStateMap.InitElevStateMap(newOrderChan)
	elevio.Init("localhost:" + port, config.NUM_FLOORS)

	currentMap := elevStateMap.GetLocalMap()
	elevStateMap.PrintMap(currentMap)
	
	init := true



    go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, orderChangesChan, newOrderChan, statusChangesChan)
    go elevio.Elevio(motorChan, doorLampChan, newOrderChan, floorChan, buttonLampChan)
	go network.Transmitter(16502, messageTx)
	go network.Receiver(16502, orderMsgRx, statusMsgRx)
    go network.PeerTransmitter(15600, id, peerTxEnable)
	go network.PeerReceiver(15600, peerUpdateCh)

    
   

	
	
	
	for {
		select {
		case p := <-peerUpdateCh:

			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)
			//currentMap := elevStateMap.GetLocalMap()
			//elevStateMap.PrintMap(currentMap)

			//initalize from network()
			if init != true{
				network.SendOrders(messageTx, elevStateMap.GetLocalMap())
			}



			



		case orderMsgFromNetwork := <- orderMsgRx:
			//Når vi mottar melding bør vi sjekke at hardware er oppdatert
			if orderMsgFromNetwork.ID != config.My_ID {
				fmt.Printf("network\n")
		
				elevStateMap.UpdateMapFromNetwork(orderMsgFromNetwork.ElevMap, newOrderChan, buttonLampChan)
				//currentMap := elevStateMap.GetLocalMap()
				//elevStateMap.PrintMap(currentMap)
				//elevStateMap.PrintMap(networkMapMsg.ElevMap)
			}
			init = false

			//currentMap := elevStateMap.GetLocalMap()
			//elevStateMap.PrintMap(currentMap)
		case statusMsgFromNetwork := <- statusMsgRx:
			if statusMsgFromNetwork.ID != config.My_ID {
				elevStateMap.UpdateElevStatusFromNetwork(statusMsgFromNetwork)
			}


		case elevMap:= <-orderChangesChan:
			fmt.Printf("gjør forandring\n")	
	
			elevStateMap.UpdateLocalMap(elevMap)
			network.SendOrders(messageTx, elevMap)
			//fmt.Printf("\n\n////////////////////DETTE SENDER JEG///////////////////\n")
			//elevStateMap.PrintMap(elevMap)
			init = false
		case elevMap:= <-statusChangesChan:
			elevStateMap.UpdateLocalMap(elevMap)
			network.SendElevStatus(messageTx, elevMap)
			init = false
		}
	}
}