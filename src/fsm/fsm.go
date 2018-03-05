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


func Fsm(motorChan chan elevio.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan elevio.ButtonLamp, mapChangesChan chan elevStateMap.ElevStateMap, newOrderChan chan elevio.ButtonEvent){
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()


	for{
	
		//checkIfAcceptOrder(mapChangesChan, buttonLampChan, ackOrderChan)
		select{
		case  floor := <- floorChan:
			fmt.Printf("NY FLOOR EVENT \n")
			eventNewFloor(motorChan, doorLampChan, doorTimer, mapChangesChan, buttonLampChan, floor)
			
		case buttonPushed := <- newOrderChan:
			

			
			fmt.Printf("En bestilling er akseptert \n")
			eventNewAckOrder(buttonLampChan, motorChan, doorLampChan, doorTimer, mapChangesChan, buttonPushed)


		case <- doorTimer.C:
			eventDoorTimeout(doorLampChan, mapChangesChan)

		}
		

	}
}


func eventNewFloor(motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, floor int){
	currentMap := elevStateMap.GetLocalMap()
	if floor != -1 {
		currentMap[config.My_ID].CurrentFloor = floor
		mapChangesChan <- currentMap

	}
	switch(state){
		case MOVING:
			if(shouldStop(currentMap)){
				motorChan <- elevio.MD_Stop
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)

				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				mapChangesChan <- currentMap
				state = DOOR_OPEN
			}
		case DOOR_OPEN:
			
	}
}

func eventDoorTimeout(doorLampChan chan bool, mapChangesChan chan elevStateMap.ElevStateMap){
	currentMap := elevStateMap.GetLocalMap()

	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			mapChangesChan <- currentMap
			state = IDLE
	}
}


func eventNewAckOrder(buttonLampChan chan elevio.ButtonLamp, motorChan chan elevio.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan elevStateMap.ElevStateMap, buttonPushed elevio.ButtonEvent){
	currentMap := elevStateMap.GetLocalMap()
	fmt.Printf("\n \n \n \n EVENT ACKORDER\n")

	buttonLampChan <- elevio.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = elevStateMap.OT_OrderPlaced
	mapChangesChan <- currentMap

	switch(state){
		case IDLE:
			if orderInThisFloor(currentMap[config.My_ID].CurrentFloor) {
				doorLampChan <- true
				
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan, mapChangesChan)
				mapChangesChan <- currentMap
				
				doorTimer.Reset(time.Second * DOOR_TIME)
				state = DOOR_OPEN
			}else{
				motorDir := chooseDirection(&currentMap)
				fmt.Printf("Skal sette motorDir %v \n", motorDir)
				mapChangesChan <- currentMap
				if motorDir != elevio.MD_Stop {
					motorChan <- motorDir
					fmt.Printf("STATE = MOVING \n")
					state = MOVING
				}
			}			
	}
}

func shouldStop(elevMap elevStateMap.ElevStateMap) bool{
	if !orderInThisFloor(elevMap[config.My_ID].CurrentFloor){
		return false
	}
	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab]==elevStateMap.OT_OrderPlaced{
		return true
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




func orderCompleted(elevMap *elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, mapChangesChan chan elevStateMap.ElevStateMap){
	fmt.Printf("ordre completed\n")

	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_Cab] = elevStateMap.OT_NoOrder
		buttonLampChan <- elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_Cab, false}
	}


	switch(elevMap[config.My_ID].CurrentDir){
		case elevStateMap.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			}
			
		case elevStateMap.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallDown] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallDown, false}
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] == elevStateMap.OT_OrderPlaced{
				elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][elevio.BT_HallUp] = elevStateMap.OT_NoOrder
				buttonLampChan <-  elevio.ButtonLamp{elevMap[config.My_ID].CurrentFloor, elevio.BT_HallUp, false}
			}
			
	}
	mapChangesChan <- *elevMap
}



func orderInThisFloor( floor int) bool{
	elevMap:= elevStateMap.GetLocalMap()
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
			fmt.Print("Min retning er opp fra før \n")
			for f := elevMap[config.My_ID].CurrentFloor+1; f < config.NUM_FLOORS; f++{
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced || nearestElevator(*elevMap, f) {
					elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
					fmt.Printf("Velger retning opp\n")
					return elevio.MD_Up
				}
			}
			for f := elevMap[config.My_ID].CurrentFloor-1; f>= 0; f-- {
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced || nearestElevator(*elevMap, f) {
					elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
					fmt.Printf("Velger retning ned\n")
					return elevio.MD_Down
				}
			}	
			break
		case elevStateMap.ED_Down:
			fmt.Printf("Min retning er ned fra før")
			for f := elevMap[config.My_ID].CurrentFloor-1; f>= 0; f-- {
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced || nearestElevator(*elevMap, f) {
					elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Down
					fmt.Printf("Velger retning ned\n")
					return elevio.MD_Down
				}
			}	
			for f := elevMap[config.My_ID].CurrentFloor+1; f < config.NUM_FLOORS; f++{
				if elevMap[config.My_ID].Orders[f][elevio.BT_Cab] == elevStateMap.OT_OrderPlaced || nearestElevator(*elevMap, f) {
					elevMap[config.My_ID].CurrentDir = elevStateMap.ED_Up
					fmt.Printf("Velger retning opp\n")
					return elevio.MD_Up
				}
			}
			break
			
	}
	return elevio.MD_Stop
}


func nearestElevator(elevMap elevStateMap.ElevStateMap, floor int) bool{

 	dist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)));
 	for e := 0; e<config.NUM_ELEVS; e++ {
 		if elevMap[e].CurrentFloor < floor && elevMap[e].CurrentDir == elevStateMap.ED_Up {
 			if elevMap[e].CurrentFloor - floor < dist {
 				return false
 			}

 		}else if elevMap[e].CurrentFloor > floor && elevMap[e].CurrentDir == elevStateMap.ED_Down {
 			if elevMap[e].CurrentFloor < floor && elevMap[e].CurrentDir == elevStateMap.ED_Up {
 				return false
 			}
 		}
 	}
 	for e := 0; e<config.NUM_ELEVS; e++ {
	 	if elevMap[e].CurrentFloor - floor == dist {
	 			return config.My_ID <= e 
	 		}
	}
 	return true
}



/*
func checkIfAcceptOrder(mapChangesChan chan elevStateMap.ElevStateMap, buttonLampChan chan elevio.ButtonLamp, ackOrderChan chan bool) {
	acceptOrder := false
		for{
			currentMap := elevStateMap.GetLocalMap()
			for f := 0; f < config.NUM_FLOORS; f++{
				for b := elevio.BT_HallUp; b < elevio.BT_Cab; b++ {
					for e := 0; e < config.NUM_ELEVS; e++ {
					
						if currentMap[e].Orders[f][b] == elevStateMap.OT_OrderExists {// && currentMap[e].Connected == true {
							acceptOrder = true
							fmt.Printf("Alle connectede heiser har 1\n")
							currentMap := elevStateMap.GetLocalMap()
							elevStateMap.PrintMap(currentMap)
						} else{
							acceptOrder = false
						}

				}
					if acceptOrder == true {

						currentMap[config.My_ID].Orders[f][b] = elevStateMap.OT_OrderAccepted
						buttonLampChan <-  elevio.ButtonLamp{f, b, true}
						mapChangesChan <- currentMap

						ackOrderChan <- true
						fmt.Printf("trigger ack order\n")
						acceptOrder = false

						
					}
				}

			}
		
		}
	}*/