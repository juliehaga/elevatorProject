package fsm

import(
	"../elevStateMap"
	"../elevio"
	"../config"
	"time"
	"math"
	"fmt"
)

var state ElevState

type ElevState int
const(
	IDLE            = 0
	MOVING          = 1
	DOOR_OPEN	    = 2
)

const DOOR_TIME 	= 2
const IDLE_TIME 	= 2


func Fsm(motorChan chan elevio.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan elevio.ButtonLamp, orderChangesChan chan elevStateMap.ElevStateMap, newOrderChan chan elevio.ButtonEvent, statusChangesChan chan elevStateMap.ElevStateMap){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()

	idleTimer := time.NewTimer(time.Second * DOOR_TIME)
	idleTimer.Stop()



	for{
		//checkIfAcceptOrder(mapChangesChan, buttonLampChan, ackOrderChan)
		select{
		case  floor := <- floorChan:
			fmt.Printf("floor event\n")
			eventNewFloor(motorChan, doorLampChan, doorTimer,orderChangesChan, buttonLampChan, floor, idleTimer, statusChangesChan)
			
		case buttonPushed := <- newOrderChan:
			fmt.Printf("buttonpushed\n")
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, orderChangesChan, buttonPushed, idleTimer, statusChangesChan)
			
		case <- doorTimer.C:
			fmt.Printf("door timeout\n")
			eventDoorTimeout(doorLampChan, statusChangesChan, idleTimer, motorChan)
			
		//case <- idleTimer.C:
		//	eventIdleTimeout(motorChan, mapChangesChan, newOrderChan, doorLampChan, doorTimer, buttonLampChan)
		//	idleTimer.Reset(time.Second * IDLE_TIME)
			

		}
		

	}
}


func eventIdleTimeout(motorChan chan elevio.MotorDirection, mapChangesChan chan elevStateMap.ElevStateMap, newOrderChan chan elevio.ButtonEvent,  doorLampChan chan bool, doorTimer *time.Timer, buttonLampChan chan elevio.ButtonLamp) {
	currentMap := elevStateMap.GetLocalMap()
	for f:= 0; f < config.NUM_FLOORS; f++{
		for b:= elevio.BT_HallUp; b <= elevio.BT_Cab; b++{
				if currentMap[config.My_ID].Orders[f][b] == elevStateMap.OT_OrderPlaced {
					if shouldStop(currentMap){
						motorChan <- elevio.MD_Stop
						doorLampChan <- true
						doorTimer.Reset(time.Second * DOOR_TIME)

						currentMap[config.My_ID].Door = true
						orderCompleted(&currentMap, buttonLampChan)
						
						currentMap[config.My_ID].IDLE = false
						state = DOOR_OPEN
						mapChangesChan <- currentMap
					} else {
						motorDir := chooseDirection(&currentMap)
						
						if motorDir != elevio.MD_Stop {
							motorChan <- motorDir
							state = MOVING
							currentMap[config.My_ID].IDLE = false

						}
						mapChangesChan <- currentMap

					
					}
				}
			
	}


		}
	}



func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, orderChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, floor int, idleTimer *time.Timer, statusChangesChan chan elevStateMap.ElevStateMap){
	currentMap := elevStateMap.GetLocalMap()
	if floor != -1 {
		currentMap[config.My_ID].CurrentFloor = floor
		statusChangesChan <- currentMap
	}
	idleTimer.Reset(time.Second * IDLE_TIME)

	switch(state){
		case MOVING:
			if shouldStop(currentMap) {
				motorChan <- elevio.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)

				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan)
				currentMap[config.My_ID].IDLE = false
				orderChangesChan <- currentMap
				state = DOOR_OPEN
			}
	}
}

func eventDoorTimeout(doorLampChan chan bool, statusChangesChan chan elevStateMap.ElevStateMap, idleTimer *time.Timer, motorChan chan elevio.MotorDirection){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			currentMap[config.My_ID].IDLE = true
			state = IDLE
			idleTimer.Reset(time.Second * IDLE_TIME)
			fmt.Printf("Door timeout")
			motorDir := chooseDirection(&currentMap)
			fmt.Printf("new direction %v\n", motorDir)
			if motorDir != elevio.MD_Stop {
				motorChan <- motorDir
				currentMap[config.My_ID].IDLE = false
				state = MOVING
			}
			statusChangesChan <- currentMap	
		
	}

	
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, orderChangesChan chan elevStateMap.ElevStateMap, buttonPushed elevio.ButtonEvent, idleTimer *time.Timer, statusChangesChan chan elevStateMap.ElevStateMap){
	fmt.Printf("evenNewAckOrder\n")

	currentMap := elevStateMap.GetLocalMap()
	buttonLampChan <- elevio.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	if buttonPushed.Button != elevio.BT_Cab{
		for elev := 0; elev < config.NUM_ELEVS; elev++{				
			currentMap[elev].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
		}
	}else {
		currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
	}	

	switch(state){
		case IDLE:

			if shouldStop(currentMap) {
				fmt.Printf("I am in ordered floor\n")
				doorLampChan <- true	
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan)
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].IDLE = false
				fmt.Printf("\n\n JEG HAR NÅ ENDRET TIL DETTE\n")
				elevStateMap.PrintMap(currentMap)
				orderChangesChan <- currentMap
				state = DOOR_OPEN

				
			}else{
				motorDir := chooseDirection(&currentMap)
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
					currentMap[config.My_ID].IDLE = false
					statusChangesChan <- currentMap
					fmt.Printf("\n\n JEG HAR NÅ ENDRET TIL DETTE\n")
					elevStateMap.PrintMap(currentMap)
					state = MOVING
				}
			}			
	}
	fmt.Printf("ut av switch\n")


	//elevStateMap.PrintMap(currentMap)
}

func shouldStop(elevMap elevStateMap.ElevStateMap) bool{
	fmt.Printf("should stop?\n")

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab]==elevStateMap.OT_OrderPlaced{
		return true
	}

	if !orderInThisFloor(elevMap[config.My_ID].CurrentFloor, elevMap){
		return false
	}

	switch elevMap[config.My_ID].CurrentDir{
		case elevStateMap.ED_Up:
			//order on current floor and no orders above
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderPlaced{
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderPlaced{
				return true }

		case elevStateMap.ED_Down:
			//order on current floor and no orders below
		 	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown]==elevStateMap.OT_OrderPlaced{
		 		return true
		 	} else if !ordersBelow(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp]==elevStateMap.OT_OrderPlaced {
			 	return true 
			}
	}
	return false
}

func ordersAbove(elevMap elevStateMap.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor + 1; f<config.NUM_FLOORS; f++{
		for b := elevio.BT_HallUp; b<= elevio.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == elevStateMap.OT_OrderPlaced{
				return true
			}
		}
	}

	return false
}

func ordersBelow(elevMap elevStateMap.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor - 1; f>=0; f--{
		for b := elevio.BT_HallUp; b<= elevio.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == elevStateMap.OT_OrderPlaced {
				return true
			}
		}
	}
	return false
}




func orderCompleted(elevMap *elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp){
	fmt.Printf("I completed an order\n")


	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] = elevStateMap.OT_NoOrder
		buttonLampChan <- elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_Cab, false}
	}


	switch(elevMap[config.My_ID].CurrentDir){
		case elevStateMap.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				
				for elev := 0; elev < config.NUM_ELEVS; elev++{				
					elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				}
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}

			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{				
					elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				}
				
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			}
			
		case elevStateMap.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{				
					elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				}



				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{				
					elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				}				
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}


			}
			
	}
}



func orderInThisFloor( floor int, elevMap elevStateMap.ElevStateMap) bool{
	for b := elevio.BT_HallUp; b <= elevio.BT_Cab; b++ {
		if elevMap[config.My_ID].Orders[floor][b] == elevStateMap.OT_OrderPlaced {
			return true
		
		}
	}
	return false
}


func chooseDirection(elevMap *elevStateMap.ElevStateMap) elevio.MotorDirection{
	fmt.Printf("Choose dir\n")
	switch elevMap[config.My_ID].CurrentDir{
		case elevStateMap.ED_Up: 
			if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if  orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						fmt.Printf("Opp fra før velger retning opp \n")
						return elevio.MD_Up
					}
				}
				
			} else if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
						fmt.Printf("Opp fra før velger retning ned \n")
						return elevio.MD_Down
					}
				}
			} else {
				return elevio.MD_Stop
			}
		case elevStateMap.ED_Down:
			if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
						fmt.Printf("Ned fra før velger retning ned \n")
						return elevio.MD_Down
					}
				}

			} else if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						fmt.Printf("Ned fra før velger retning opp \n")
						return elevio.MD_Up
					}
				}
			} else {
				return elevio.MD_Stop
			}
	}
	fmt.Printf("jeg blir stånde\n")
	return elevio.MD_Stop
}



func nearestElevator(elevMap elevStateMap.ElevStateMap, floor int) bool{

 	myDist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)))



 	if elevMap[config.My_ID].CurrentFloor < floor { 
 		fmt.Printf("jeg står under \n")
	 	for e := 0; e<config.NUM_ELEVS; e++ {
		 	if e != config.My_ID{	
		 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
		 		fmt.Printf("Den andre er %v etasjer unna\n", distElev)
		 		if distElev < myDist{
		 			fmt.Printf("den har kortere\n")
		 			if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE ){
		 				fmt.Printf("den andre tar den\n")
		 				return false
		 			} else if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE ) {
		 				fmt.Printf("den andre tar den\n")
		 				return false
		 			}
		 		} else if myDist == distElev && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE){
		 			if e > config.My_ID{
		 				return false
		 			}
		 		}
		 	}
		 }
 	} else if elevMap[config.My_ID].CurrentFloor > floor {
	 		fmt.Printf("jeg står over \n")
		 	for e := 0; e<config.NUM_ELEVS; e++ {
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			fmt.Printf("den har kortere\n")
			 			if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE ){
			 				fmt.Printf("den andre tar den\n")
			 				return false
			 			} else if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE ) {
		 					fmt.Printf("den andre tar den\n")
		 					return false
		 				}
			 		}  else if myDist == distElev && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE){
		 			if e > config.My_ID{
		 				fmt.Printf("Prioritert på ID")
		 				return false
		 			}
		 		}
			 	}
			}
	}

	fmt.Printf("jeg var nærmest\n")
 	return true
}
