package fsm

import(
	"../elevStateMap"
	"../config"
	"time"
	"math"
	"fmt"
)

var state ElevState

type ElevState int
const(
	INIT            = 0
	IDLE            = 1
	MOVING          = 2
	DOOR_OPEN	    = 3
	OUT_OF_ORDER    = 4
)

const DOOR_TIME 	    = 2
const IDLE_TIME 	    = 9
const MOTOR_DEAD_TIME 	= 10


func Fsm(motorChan chan config.MotorDirection, doorLampChan chan bool, floorChan chan int, buttonLampChan chan config.ButtonLamp, mapChangesChan chan config.ElevStateMap, newOrderChan chan config.ButtonEvent, statusChangesChan chan config.ElevStateMap, orderCompleteChan chan config.ButtonEvent){
	state = INIT
	
	doorTimer := time.NewTimer(time.Second * DOOR_TIME)
	doorTimer.Stop()

	idleTimer := time.NewTimer(time.Second * IDLE_TIME)
	idleTimer.Stop()

	motorTimer := time.NewTimer(time.Second * MOTOR_DEAD_TIME)
	motorTimer.Stop()

	


	for{
		
		select{
		case  floor := <- floorChan:
			//fmt.Printf("floor event\n")
			eventNewFloor(orderCompleteChan, motorChan, doorLampChan, doorTimer,mapChangesChan, buttonLampChan, floor, idleTimer, statusChangesChan, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)
			motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
			//fmt.Printf("motor reset %v\n", bool)

		case buttonPushed := <- newOrderChan:
			//fmt.Printf("buttonpushed\n")
			eventNewAckOrder(orderCompleteChan, buttonLampChan, motorChan, doorLampChan, doorTimer, mapChangesChan, buttonPushed, idleTimer, motorTimer, statusChangesChan)
			idleTimer.Reset(time.Second * IDLE_TIME)

		case <- doorTimer.C:
			//fmt.Printf("door timeout\n")
			eventDoorTimeout(doorLampChan, mapChangesChan, idleTimer, motorChan, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)
			
		/*case <- idleTimer.C:
			eventIdleTimeout(motorChan, statusChangesChan, mapChangesChan, doorLampChan, doorTimer, buttonLampChan, motorTimer)
			idleTimer.Reset(time.Second * IDLE_TIME)*/
/*
		case <- motorTimer.C:
			currentMap := elevStateMap.GetLocalMap()
			//fmt.Printf("IDLE = %v", currentMap[config.My_ID].IDLE)
			if (currentMap[config.My_ID].IDLE == false){
				fmt.Printf("I'm out of order\n\n")
				currentMap[config.My_ID].OutOfOrder = true
				eventOutOfOrder(motorChan, statusChangesChan)
				motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
				fmt.Printf("motor reset %v\n", bool)
				state = OUT_OF_ORDER
				mapChangesChan <- currentMap
			}
			*/
		}
	}
}

func eventOutOfOrder(motorChan chan config.MotorDirection, statusChangesChan chan config.ElevStateMap){
	currentMap := elevStateMap.GetLocalMap()
	//fmt.Printf("Out of order\n")
	if currentMap[config.My_ID].CurrentFloor != config.NUM_FLOORS -1{
		motorChan <- config.MD_Up
		currentMap[config.My_ID].CurrentDir = config.ED_Up
	} else {
		motorChan <- config.MD_Down
		currentMap[config.My_ID].CurrentDir = config.ED_Down
	}
}

/*
func eventIdleTimeout(motorChan chan config.MotorDirection, statusChangesChan chan config.ElevStateMap, mapChangesChan chan config.ElevStateMap, doorLampChan chan bool, doorTimer *time.Timer, buttonLampChan chan config.ButtonLamp,  motorTimer *time.Timer) {
	//fmt.Printf("In idle timeout")
	currentMap := elevStateMap.GetLocalMap()
	motorDir := forceChooseDirection(&currentMap, motorTimer)
	if motorDir != config.MD_Stop {
		motorChan <- motorDir
		currentMap[config.My_ID].IDLE = false
		state = MOVING

	} else if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
		doorLampChan <- true
		doorTimer.Reset(time.Second * DOOR_TIME)

		currentMap[config.My_ID].Door = true
		currentMap = orderCompleted(currentMap, buttonLampChan)
		currentMap[config.My_ID].IDLE = false
		state = DOOR_OPEN
	}
	mapChangesChan <- currentMap
}

*/

func eventNewFloor(orderCompleteChan chan config.ButtonEvent, motorChan chan config.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan config.ElevStateMap, buttonLampChan chan config.ButtonLamp, floor int, idleTimer *time.Timer, statusChangesChan chan config.ElevStateMap, motorTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	//var det en grunn til at vi skulle oppdatere currentfloor her? 
	currentMap[config.My_ID].CurrentFloor = floor 
	var motorDir config.MotorDirection
	switch(state){
		case INIT:
			motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap, motorTimer)
			if motorDir == config.MD_Stop{
				state = IDLE
			} else{
				motorChan <- motorDir
				state = MOVING
			}
			
		case MOVING:
			//motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
			if shouldStop(currentMap) {
				motorChan <- config.MD_Stop
					if  orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
						doorLampChan <- true
						doorTimer.Reset(time.Second * DOOR_TIME)
						currentMap[config.My_ID].Door = true
						currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan)
						currentMap[config.My_ID].IDLE = false
						elevStateMap.PrintMap(currentMap)
						state = DOOR_OPEN
					} else {
						currentMap[config.My_ID].IDLE = true
						state = IDLE
					}	
			}
		/*
		case OUT_OF_ORDER:
			if shouldStop(currentMap){
				motorChan <- config.MD_Stop
				if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
					doorLampChan <- true
					doorTimer.Reset(time.Second * DOOR_TIME)
					currentMap[config.My_ID].Door = true
					currentMap = orderCompleted(currentMap, buttonLampChan)
					currentMap[config.My_ID].IDLE = false
					state = DOOR_OPEN
				} else{
					state = IDLE
					currentMap[config.My_ID].IDLE = true
				}
			} else {
				motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap, motorTimer)
				if motorDir != config.MD_Stop {
					motorChan <- motorDir
					currentMap[config.My_ID].IDLE = false
					state = MOVING
				} else{
					state = IDLE
					currentMap[config.My_ID].IDLE = true
				}
				currentMap[config.My_ID].OutOfOrder = false //Skal denne være her???
			}*/

	}

		//hvorfor bruker vi ikke shouldStop her??

		/*	
		case OUT_OF_ORDER:
			if  orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
				doorLampChan <- true
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].Door = true
				orderCompleted(&currentMap, buttonLampChan)
				motorChan <- config.MD_Stop
				currentMap[config.My_ID].IDLE = false
				state = DOOR_OPEN
				mapChangesChan <- currentMap //Denne trenger ikke være her??
			}else if (currentMap[config.My_ID].CurrentDir == config.ED_Up && !ordersAbove(currentMap)) || (currentMap[config.My_ID].CurrentDir == config.ED_Down && !ordersBelow(currentMap)) {
				motorChan <- config.MD_Stop
				state = IDLE
				currentMap[config.My_ID].IDLE = true
			}else{
				motorDir := chooseDirection(&currentMap, motorTimer)
				if motorDir != config.MD_Stop {
					motorChan <- motorDir
					currentMap[config.My_ID].IDLE = false
					state = MOVING
				} else{
					state = IDLE
					currentMap[config.My_ID].IDLE = true
				}
			currentMap[config.My_ID].OutOfOrder = false //Skal denne være her???

			}
		}*/
	mapChangesChan <- currentMap
		
}

func eventDoorTimeout(doorLampChan chan bool, mapChangesChan chan config.ElevStateMap, idleTimer *time.Timer, motorChan chan config.MotorDirection, motorTimer *time.Timer){
	currentMap := elevStateMap.GetLocalMap()
	var motorDir config.MotorDirection
	fmt.Printf("*******************DOOR TIMEOUT**************\n")
	elevStateMap.PrintMap(currentMap)
	switch(state){
		case DOOR_OPEN:
			doorLampChan <- false
			currentMap[config.My_ID].Door = false
			motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap, motorTimer)
			motorChan <- motorDir
			if motorDir != config.MD_Stop {
				currentMap[config.My_ID].IDLE = false
				state = MOVING
			} else {
				//motorTimer.Stop()
				currentMap[config.My_ID].IDLE = true
				state = IDLE
				
			}
			mapChangesChan <- currentMap	
	}
}


func eventNewAckOrder(orderCompleteChan chan config.ButtonEvent, buttonLampChan chan config.ButtonLamp, motorChan chan config.MotorDirection, doorLampChan chan bool, doorTimer *time.Timer, mapChangesChan chan config.ElevStateMap, buttonPushed config.ButtonEvent, idleTimer *time.Timer, motorTimer *time.Timer, statusChangesChan chan config.ElevStateMap){
	currentMap := elevStateMap.GetLocalMap()
	//accept CAB order
	if buttonPushed.Button == config.BT_Cab{ //&& currentMap[config.My_ID].OutOfOrder == false{
		currentMap[config.My_ID].Orders[buttonPushed.Floor][buttonPushed.Button] = config.OT_OrderPlaced
		buttonLampChan <- config.ButtonLamp{buttonPushed.Floor, buttonPushed.Button, true}
	}
		

	
	var motorDir config.MotorDirection
	switch(state){
		case IDLE:
			if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){// && currentMap[config.My_ID].OutOfOrder == false{
				//endret status og ordre
				//fmt.Printf("order, in this floor\n")
				doorLampChan <- true	
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].Door = true
				currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan)
				currentMap[config.My_ID].IDLE = false
				state = DOOR_OPEN
				time.Sleep(2000*time.Millisecond)
			}else{
				//fmt.Printf("Jeg har lyst til å velge retning \n")
				motorDir, currentMap[config.My_ID].CurrentDir = chooseDirection(currentMap, motorTimer)
				motorChan <- motorDir
				if motorDir != config.MD_Stop {
					state = MOVING
					currentMap[config.My_ID].IDLE = false
				} else{
					//motorTimer.Stop()
				}

			}
		case DOOR_OPEN:
			if orderInThisFloor(currentMap[config.My_ID].CurrentFloor, currentMap){
				//fmt.Printf("door open order, in this floor\n")
				doorLampChan <- true	
				doorTimer.Reset(time.Second * DOOR_TIME)
				currentMap[config.My_ID].Door = true
				currentMap = orderCompleted(currentMap, buttonLampChan, orderCompleteChan)
				currentMap[config.My_ID].IDLE = false
				time.Sleep(2000*time.Millisecond)
			}
	}
	mapChangesChan <- currentMap
}



func shouldStop(elevMap config.ElevStateMap) bool{
	switch(state){
		case MOVING: 
			if elevMap[config.My_ID].CurrentFloor == config.NUM_FLOORS-1 && elevMap[config.My_ID].CurrentDir == config.ED_Up{
				return true
			} else if elevMap[config.My_ID].CurrentFloor == 0 && elevMap[config.My_ID].CurrentDir == config.ED_Down {
				return true
			}
	}
	
	//elevStateMap.PrintMap(elevMap)

	//Er BT_cab alltid bare local order??
	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab]==config.OT_OrderPlaced {
		return true
	}

	switch elevMap[config.My_ID].CurrentDir{
		case config.ED_Up:
			//order on current floor and no orders above
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp]==config.OT_OrderPlaced {
				return true
			} else if !ordersAbove(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown]==config.OT_OrderPlaced{
				return true 
			} else if !ordersAbove(elevMap){
				return true
			}

		case config.ED_Down:
			//order on current floor and no orders below
		 	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown]==config.OT_OrderPlaced {
		 		return true
		 	} else if !ordersBelow(elevMap) && elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp]==config.OT_OrderPlaced {
			 	return true 
			} else if !ordersBelow(elevMap){
				return true
			}
	}
	return false
}

func ordersAbove(elevMap config.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor + 1; f<config.NUM_FLOORS; f++{
		for b := config.BT_HallUp; b <= config.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced{


				return true
			}
		}
	}

	return false
}

func ordersBelow(elevMap config.ElevStateMap) bool{
	for f := elevMap[config.My_ID].CurrentFloor - 1; f>=0; f--{
		for b := config.BT_HallUp; b<= config.BT_Cab; b++{ 
			if elevMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced{
				return true
			}
		}
	}
	return false
}



func orderCompleted(elevMap config.ElevStateMap, buttonLampChan chan config.ButtonLamp, orderCompleteChan chan config.ButtonEvent) config.ElevStateMap{
	fmt.Printf("ordercompleted\n")
	if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab] == config.OT_OrderPlaced{
		elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_Cab] = config.OT_NoOrder
		//fmt.Printf("Completed CAB order\n")
		buttonLampChan <- config.ButtonLamp{elevMap[config.My_ID].CurrentFloor, config.BT_Cab, false}
	}

	switch(elevMap[config.My_ID].CurrentDir){
		case config.ED_Up: 
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] == config.OT_OrderPlaced{
				
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallUp}, buttonLampChan)

			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] == config.OT_OrderPlaced {
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallDown}, buttonLampChan)		
			}
			
		case config.ED_Down:
			if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallDown] == config.OT_OrderPlaced{
				
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallDown}, buttonLampChan)				
			
			} else if elevMap[config.My_ID].Orders[elevMap[config.My_ID].CurrentFloor][config.BT_HallUp] == config.OT_OrderPlaced{
				
				elevMap = ClearOrder(elevMap, config.ButtonEvent{elevMap[config.My_ID].CurrentFloor, config.BT_HallUp}, buttonLampChan)
			}
			
	}
	return elevMap
}




func orderInThisFloor( floor int, elevMap config.ElevStateMap) bool{
	//elevStateMap.PrintMap(elevMap)
	if (floor != -1){
		for b := config.BT_HallUp; b <= config.BT_Cab; b++ {
			if elevMap[config.My_ID].Orders[floor][b] == config.OT_OrderPlaced{
				return true
			}
		}
	}
	return false
}


func chooseDirection(elevMap config.ElevStateMap, motorTimer *time.Timer) (config.MotorDirection, config.ElevDir){
	//fmt.Printf("choose dir\n")
	motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
	//fmt.Printf("motor reset %v\n", bool)
	switch elevMap[config.My_ID].CurrentDir{
		case config.ED_Up: 
			if ordersAbove(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if  orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						//fmt.Printf("velger opp\n")
						return config.MD_Up, config.ED_Up
					}
				}
			} else if ordersBelow(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						//fmt.Printf("Velger ned \n")
						return config.MD_Down, config.ED_Down
					}
				}
			}	
		case config.ED_Down:
			if ordersBelow(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor - 1; f >= 0; f--{
					if orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced){
						//fmt.Printf("Velger ned \n")
						return config.MD_Down, config.ED_Down
					}
				}
			} else if ordersAbove(elevMap){
				for f:= elevMap[config.My_ID].CurrentFloor + 1; f < config.NUM_FLOORS; f++{
					if orderInThisFloor(f, elevMap) && (nearestElevator(elevMap, f) || elevMap[config.My_ID].Orders[f][config.BT_Cab] == config.OT_OrderPlaced ){
						//fmt.Printf("velger opp\n")
						return config.MD_Up, config.ED_Up
					}
				}
			}
	}
	//fmt.Printf("STOP\n")
	return config.MD_Stop, elevMap[config.My_ID].CurrentDir
}



func nearestElevator(elevMap config.ElevStateMap, floor int) bool{

 	myDist := int(math.Abs(float64(elevMap[config.My_ID].CurrentFloor - floor)))
 	//fmt.Printf("*********************************************\n")
 	//elevStateMap.PrintMap(elevMap)
 	//fmt.Printf("*********************************************\n")


 	if elevMap[config.My_ID].CurrentFloor < floor { 
	 	for e := 0; e<config.NUM_ELEVS; e++ {
	 		if elevMap[e].Connected == true{ //&& elevMap[e].OutOfOrder == false{	
			 	if e != config.My_ID{	
			 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
			 		if distElev < myDist{
			 			if elevMap[e].CurrentFloor < floor && (elevMap[e].CurrentDir == config.ED_Up || elevMap[e].IDLE ){
			 				return false
			 			} else if elevMap[e].CurrentFloor > floor && (elevMap[e].CurrentDir == config.ED_Down || elevMap[e].IDLE ) {
			 				return false
			 			} else if  elevMap[e].CurrentFloor == floor && elevMap[e].IDLE {
			 				return false
			 			}
			 		} else if myDist == distElev && (elevMap[e].CurrentDir == config.ED_Up || elevMap[e].IDLE){
			 			if e > config.My_ID{
			 				return false
			 			}
			 		}
			 	}
			} 	
		 }
 	} else if elevMap[config.My_ID].CurrentFloor > floor {
		 	for e := 0; e<config.NUM_ELEVS; e++ {
		 		if elevMap[e].Connected == true{ //&& elevMap[e].OutOfOrder == false{	
				 	if e != config.My_ID{	
				 		distElev := int(math.Abs(float64(elevMap[e].CurrentFloor - floor)))
				 		if distElev < myDist{
				 			if elevMap[e].CurrentFloor >= floor && (elevMap[e].CurrentDir ==config.ED_Down || elevMap[e].IDLE ){
				 				return false
				 			} else if elevMap[e].CurrentFloor <= floor && (elevMap[e].CurrentDir == config.ED_Up || elevMap[e].IDLE ) {
			 					return false
			 				} else if  elevMap[e].CurrentFloor == floor && elevMap[e].IDLE {
			 					return false
				 			} 

				 		}else if myDist == distElev && (elevMap[e].CurrentDir == config.ED_Down || elevMap[e].IDLE){
			 			if e > config.My_ID{
			 				return false
			 			}
			 		}
			 	}
			}
		}	
	}
	//fmt.Printf("Jeg er nærmest\n")
 	return true
}


//Trenger ikke denne funksjonen om vi ikke skal ha motorTimeout
func forceChooseDirection(elevMap *config.ElevStateMap, motorTimer *time.Timer) config.MotorDirection{
	motorTimer.Reset(time.Second * MOTOR_DEAD_TIME)
	//fmt.Printf("motor reset %v\n", bool)
	elevsInIdle := 0
	for e := 0; e < config.NUM_ELEVS; e++{
		if e != config.My_ID{
			if elevMap[e].IDLE == true{
				elevsInIdle ++
			}
		}
	}
	if orderInThisFloor(elevMap[config.My_ID].CurrentFloor, *elevMap){
		return config.MD_Stop
	
	} else if elevMap[config.My_ID].CurrentFloor >= config.NUM_FLOORS/2 {
		if ordersAbove(*elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Up
			return config.MD_Up
		} else if ordersBelow(*elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Down
			return config.MD_Down
		}
	} else if elevMap[config.My_ID].CurrentFloor < config.NUM_FLOORS/2{
		if ordersBelow(*elevMap) {
			elevMap[config.My_ID].CurrentDir = config.ED_Down
			return config.MD_Down
		} else if ordersAbove(*elevMap){
			elevMap[config.My_ID].CurrentDir = config.ED_Up
			return config.MD_Up
		}
	}
	return config.MD_Stop
} 


//Kanskje vi må sende med map her, mulig det blir feil å bruke den i completedOrder om ikke???
func ClearOrder(elevMap config.ElevStateMap, button config.ButtonEvent, buttonLampChan chan config.ButtonLamp) config.ElevStateMap{

	for elev := 0; elev < config.NUM_ELEVS; elev++{				
		elevMap[elev].Orders[button.Floor][button.Button] = config.OT_NoOrder
	}	
				//fmt.Printf("completed HALLUP %v\n", elevMap[config.My_ID].CurrentFloor)
	buttonLampChan <-  config.ButtonLamp{button.Floor, button.Button, false}
	return elevMap
}