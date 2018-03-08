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
			idleTimer.Reset(time.Second * IDLE_TIME)

		case buttonPushed := <- newOrderChan:
			fmt.Printf("buttonpushed\n")
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, orderChangesChan, buttonPushed, idleTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case <- doorTimer.C:
			fmt.Printf("door timeout\n")
			eventDoorTimeout(doorLampChan, statusChangesChan, idleTimer, motorChan)
			idleTimer.Reset(time.Second * IDLE_TIME)
			
		case <- idleTimer.C:
			eventIdleTimeout(motorChan, statusChangesChan)
			idleTimer.Reset(time.Second * IDLE_TIME)
			

		}
		

	}
}


func eventIdleTimeout(motorChan chan elevio.MotorDirection, statusChangesChan chan elevStateMap.ElevStateMap) {
	currentMap := elevStateMap.GetLocalMap()
	motorDir := chooseDirection(&currentMap)
		if motorDir != elevio.MD_Stop {
			motorChan <- motorDir
			currentMap[config.My_ID].IDLE = false
			state = MOVING
		}
	statusChangesChan <- currentMap
}



func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, orderChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, floor int, idleTimer *time.Timer, statusChangesChan chan elevStateMap.ElevStateMap){
	currentMap := elevStateMap.GetLocalMap()
	if floor != -1 {
		currentMap[config.My_ID].CurrentFloor = floor
		statusChangesChan <- currentMap
	}

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
	//idleTimer.Reset(time.Second * IDLE_TIME)
}

func eventDoorTimeout(doorLampChan chan bool, statusChangesChan chan elevStateMap.ElevStateMap, idleTimer *time.Timer, motorChan chan elevio.MotorDirection){
	currentMap := elevStateMap.GetLocalMap()
	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			currentMap[config.My_ID].IDLE = true
			state = IDLE
			//idleTimer.Reset(time.Second * IDLE_TIME)

			motorDir := chooseDirection(&currentMap)
			if motorDir != elevio.MD_Stop {
				motorChan <- motorDir
				currentMap[config.My_ID].IDLE = false
				state = MOVING
			}
			statusChangesChan <- currentMap	
		
	}

	
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, orderChangesChan chan elevStateMap.ElevStateMap, buttonPushed elevio.ButtonEvent, idleTimer *time.Timer){

	currentMap := elevStateMap.GetLocalMap()
	fmt.Printf("\n \n CURRENT FLOOR %v \n \n", currentMap[config.My_ID].CurrentFloor)
	buttonLampChan <- elevio.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	if buttonPushed.Button != elevio.BT_Cab{
		for elev := 0; elev < config.NUM_ELEVS; elev++{		
			if currentMap[elev].Connected == true{		
				currentMap[elev].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
			}
		}	
	}else {
		currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
	}	

	switch(state){
		case IDLE:

			if shouldStop(currentMap) {
				fmt.Printf("IDLE -shouldstop\n")
				doorLampChan <- true	
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan)
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].IDLE = false
				state = DOOR_OPEN

				
			}else{
				motorDir := chooseDirection(&currentMap)
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
					currentMap[config.My_ID].IDLE = false
					
					state = MOVING
				}

				
			}	

	}
	orderChangesChan <- currentMap
	//idleTimer.Reset(time.Second * IDLE_TIME)


	//elevStateMap.PrintMap(currentMap)
}

func shouldStop(elevMap elevStateMap.ElevStateMap) bool{
	switch(state){
		case MOVING: 
			fmt.Printf("state moving\n")
			if elevMap[config.My_ID].CurrentFloor == config.NUM_FLOORS && elevMap[config.My_ID].CurrentDir == elevStateMap.ED_Up{
				return true
			} else if elevMap[config.My_ID].CurrentFloor == 0 && elevMap[config.My_ID].CurrentDir == elevStateMap.ED_Down {
				return true
			}
	}
	
	//elevStateMap.PrintMap(elevMap)

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

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] = elevStateMap.OT_NoOrder
		buttonLampChan <- elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_Cab, false}
	}


	switch(elevMap[config.My_ID].CurrentDir){
		case elevStateMap.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				
				for elev := 0; elev < config.NUM_ELEVS; elev++{	
					if elevMap[elev].Connected == true{				
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
					}
				}	
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}

			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{			
					if elevMap[elev].Connected == true{		
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
					}
				}
				
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			}
			
		case elevStateMap.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{		
					if elevMap[elev].Connected == true{			
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
					}
				}


				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{

				for elev := 0; elev < config.NUM_ELEVS; elev++{	
					if elevMap[elev].Connected == true{				
						elevMap[elev].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
					}
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
	switch elevMap[config.My_ID].CurrentDir{
		case elevStateMap.ED_Up: 
			if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if  orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						return elevio.MD_Up
					}
				}
				
			} else if ordersBelow(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
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
						return elevio.MD_Down
					}
				}

			} else if ordersAbove(*elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if orderInThisFloor(f, *elevMap) && (nearestElevator(*elevMap, f) || elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced){
						elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
						return elevio.MD_Up
					}
				}
			} else {
				return elevio.MD_Stop
			}
	}
	return elevio.MD_Stop
}



func nearestElevator(elevMap elevStateMap.ElevStateMap, floor int) bool{

 	myDist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)))



 	if elevMap[config.My_ID].CurrentFloor < floor { 
	 	for e := 0; e<config.NUM_ELEVS; e++ {
	 		if elevMap[e].Connected == true{	
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE ){
			 				return false
			 			} else if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE ) {
			 				return false
			 			}
			 		} else if myDist == distElev && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE){
			 			if e > config.My_ID{
			 				return false
			 			}
			 		}
			 	}
			} 	
		 }
 	} else if elevMap[config.My_ID].CurrentFloor > floor {
		 	for e := 0; e<config.NUM_ELEVS; e++ {
		 		if elevMap[e].Connected == true{	
				 	if e != config.My_ID{	
				 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
				 		if distElev < myDist{
				 			if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE ){
				 				return false
				 			} else if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == elevStateMap.ED_Up || elevMap[e].IDLE ) {
			 					return false
			 				}
				 		}  else if myDist == distElev && (elevMap[e].CurrentDir == elevStateMap.ED_Down || elevMap[e].IDLE){
			 			if e > config.My_ID{
			 				return false
			 			}
			 		}
			 	}
			}
		}	
	}
 	return true
}
