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
    newLocalOrderChan := make(chan config.ButtonEvent, 100)
  
    orderMsgChan := make(chan config.ElevStateMap, 100)

    // We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
    peerUpdateCh := make(chan config.PeerUpdate, 100)
    // This could be used to signal that we are somehow "unavailable".
    peerTxEnable := make(chan bool, 100)

 
	messageTx := make(chan config.Message, 10000)
	orderMsgRx := make(chan config.OrderMsg, 10000)
	statusMsgRx := make(chan config.StatusMsg, 10000)
	ackChan := make(chan config.AckMsg, 10000)
	orderCompleteChan := make(chan config.ButtonEvent, 10000)
	
	config.Init(id, port)
	elevio.InitDriver("localhost:" + port, config.NUM_FLOORS)
	elevStateMap.InitElevStateMap()
	elevio.InitOrders()

	fmt.Printf("Init success\n")
	
	init := true



  
    go elevio.Elevio(motorChan, doorLampChan, newOrderChan, floorChan, buttonLampChan, orderMsgChan, newLocalOrderChan, mapChangesChan)
	go network.Transmitter(16502, messageTx, ackChan)
	go network.Receiver(16502, orderMsgRx, statusMsgRx, ackChan, messageTx)
    go network.PeerTransmitter(15600, id, peerTxEnable)
	go network.PeerReceiver(15600, peerUpdateCh)
	go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, mapChangesChan, newOrderChan, statusChangesChan, orderCompleteChan)

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
			fmt.Printf("Får melding fra %v", orderMsgFromNetwork.ID)

			orderUpdates, currentMap := elevStateMap.UpdateMapFromNetwork(orderMsgFromNetwork.ElevMap, buttonLampChan)
			if init == true{
				elevio.InitOrders()
			}
			init = false

			if orderUpdates {
				orderMsgChan <- currentMap
				fmt.Printf("//////////// LEGGER PÅ ORDERMSGCHAN FRA NETTVERK/////////////////////////\n")
				elevStateMap.PrintMap(currentMap)
				network.SendOrders(messageTx, currentMap)
			}

		case statusMsgFromNetwork := <- statusMsgRx:
			elevStateMap.UpdateElevStatusFromNetwork(statusMsgFromNetwork)
			

		case elevMap:= <-mapChangesChan:
			localOrderUpdates := elevStateMap.UpdateLocalMap(elevMap)
			if localOrderUpdates {
				orderMsgChan <- elevMap
				//fmt.Printf("//////////// LEGGER PÅ ORDERMSGCHAN LOKALT/////////////////////////\n")
				//elevStateMap.PrintMap(elevStateMap.GetLocalMap())
				network.SendOrders(messageTx, elevMap)
				//elevStateMap.PrintMap(elevMap)
			}
			network.SendElevStatus(messageTx, elevMap)
			init = false










			//add HALL order
		
	

		/*case button:= <- orderCompleteChan:
			fmt.Printf("I completed an order, sending msg\n")
			//elevStateMap.PrintMap(elevStateMap.GetLocalMap())
			network.SendOrderComplete(messageTx, button)

		case order := <- clearOrderChan:
			elevMap := elevStateMap.GetLocalMap()
			fmt.Printf("msg from network about clear order\n")
			
			elevMap = fsm.ClearOrder(elevMap, order, buttonLampChan)
			//elevStateMap.PrintMap(elevMap)*/


/*		case elevMap:= <-statusChangesChan:
			elevStateMap.UpdateLocalMap(elevMap)
			network.SendElevStatus(messageTx, elevMap)
			init = false*/
		}
	}
}