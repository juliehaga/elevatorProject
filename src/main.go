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


	
	
//channels for communication between module


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

	activeOrderRx := make(chan config.ActiveOrders, 10000)
	activeOrderTx := make(chan config.ActiveOrders, 10000)
	
	
	config.Init(id, port)
	elevio.InitDriver("localhost:" + port, config.NUM_FLOORS)
	elevStateMap.InitElevStateMap()
	elevio.InitOrders()

	fmt.Printf("Init success\n")
	
	init := true
	ActiveOrderMatrix := [config.NUM_FLOORS][config.NUM_BUTTONS][config.NUM_ELEVS] bool{}



  
    go elevio.Elevio(motorChan, doorLampChan, newOrderChan, floorChan, buttonLampChan, orderMsgChan, newLocalOrderChan, mapChangesChan)
	go network.Transmitter(16666, messageTx, ackChan)
	go network.Receiver(16666, orderMsgRx, statusMsgRx, ackChan, messageTx, activeOrderRx)
    go network.PeerTransmitter(15611, id, peerTxEnable)
	go network.PeerReceiver(15611, peerUpdateCh)
	go fsm.Fsm(motorChan, doorLampChan, floorChan, buttonLampChan, mapChangesChan, newOrderChan, statusChangesChan, orderCompleteChan, activeOrderTx)
	go elevStateMap.FindActiveOrders(orderMsgChan, activeOrderTx)

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
			fmt.Printf("Får melding fra %v\n", orderMsgFromNetwork.ID)

			orderUpdates, currentMap := elevStateMap.UpdateMapFromNetwork(orderMsgFromNetwork.ElevMap, buttonLampChan)
			if init == true{
				elevio.InitOrders()
			}
			init = false

			if orderUpdates {
				orderMsgChan <- currentMap
				fmt.Printf("//////////// Sender mine ordre, NETWORK /////////////////////////\n")
				//elevStateMap.PrintMap(currentMap)
				network.SendOrders(messageTx, currentMap)
			}

		case statusMsgFromNetwork := <- statusMsgRx:
			elevStateMap.UpdateElevStatusFromNetwork(statusMsgFromNetwork)
			

		case elevMap:= <-mapChangesChan:
			localOrderUpdates := elevStateMap.UpdateLocalMap(elevMap)
			if localOrderUpdates {
				orderMsgChan <- elevMap
				fmt.Printf("//////////// Sender mine ordre, LOCAL endring/////////////////////////\n")
				//elevStateMap.PrintMap(elevStateMap.GetLocalMap())
				network.SendOrders(messageTx, elevMap)
				//elevStateMap.PrintMap(elevMap)
			}
			network.SendElevStatus(messageTx, elevMap)
			init = false

		case order:= <- activeOrderTx:
			//sjekk om den skal aktiveres eller cleares

			if order.ActiveOrder {
				fmt.Printf("Jeg sender en aktiv ordreMSG\n")
				ActiveOrderMatrix[order.Button.Floor][order.Button.Button][config.My_ID] = true
				network.SendActiveOrder(messageTx, order)

			} else {
				for e:= 0; e < config.NUM_ELEVS; e++{
					ActiveOrderMatrix[order.Button.Floor][order.Button.Button][e] = false
				}
			}

		case order:= <- activeOrderRx:

			
			ActiveOrderMatrix[order.Button.Floor][order.Button.Button][order.ID] = true
			fmt.Printf("ORDRE MELDING FRA %v\n", order.ID)
			fmt.Printf("%v\n", ActiveOrderMatrix)
			newOrder := true
			for e := 0; e < config.NUM_ELEVS; e++{
				if ActiveOrderMatrix[order.Button.Floor][order.Button.Button][e] == false {
					newOrder = false
				}
			}

			if newOrder{
				fmt.Printf("trigger new order chan\n")
				newOrderChan <- config.ButtonEvent{order.Button.Floor, order.Button.Button}
				buttonLampChan <- config.ButtonLamp{order.Button.Floor, order.Button.Button, true}
			}
			











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