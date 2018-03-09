package elevStateMap

import(
	"../config"
	"fmt"
	"sync"
	"os"
	"strconv"
	"io"
)


var LocalMap config.ElevStateMap
var mutex = &sync.Mutex{}




func writeToBackup(){
	// open file using READ & WRITE permission
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()

	// write some text line-by-line to file

	for f := 0; f< config.NUM_FLOORS; f++{
		var order int = int(LocalMap[config.My_ID].Orders[f][config.BT_Cab])
		_, err = file.WriteString(strconv.Itoa(order))
		if err != nil { return }
	}
	// save changes
	err = file.Sync()
	if err != nil { return }
}

func readFromBackup(newOrderChan chan config.ButtonEvent){
	// re-open file
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()
	currentMap := GetLocalMap()

	// read file, line by line

	var buf = make([]byte, 1024)
	for {
		_, err = file.Read(buf)
		
		// break if finally arrived at end of file
		if err == io.EOF {
			break
		}
		
		// break if error occured
		if err != nil && err != io.EOF {
			break
		}
		//order := strconv.Atoi((string(buf)))

		fmt.Printf("order 1: %v\n", string(buf[0]))
		fmt.Printf("order 1: %v\n", string(buf[1]))
		fmt.Printf("order 1: %v\n", string(buf[2]))
		fmt.Printf("order 1: %v\n", string(buf[3]))


		//fmt.Printf("New line %v \n", order)
	}
	for floor := 0; floor<config.NUM_FLOORS; floor++{
		order, _ :=strconv.Atoi(string(buf[floor]))
		currentMap[config.My_ID].Orders[floor][config.BT_Cab] = config.OrderType(order)
	}

	SetLocalMap(currentMap)
}

func InitElevStateMap(newOrderChan chan config.ButtonEvent){
	var _, err = os.Stat("backup.txt")

	if os.IsNotExist(err) {
		fmt.Printf("Creating new backup file\n")
		var file, err = os.Create("backup.txt")
		if err != nil{
			fmt.Printf("Could not create file \n")
			//return
		}	
		for f := 0; f < config.NUM_FLOORS; f++{
			LocalMap[config.My_ID].Orders[f][config.BT_Cab] = config.OT_NoOrder
		}
		defer file.Close()
	}else{
		readFromBackup(newOrderChan)
	}


	//initialize map	
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].IDLE = true
		LocalMap[e].CurrentDir = config.ED_Down
		LocalMap[e].Connected = false
		LocalMap[e].Door = false
		LocalMap[e].CurrentFloor = -1
		LocalMap[e].OutOfOrder = false
			
		for f := 0; f < config.NUM_FLOORS; f++{
			for b :=0; b < config.BT_Cab; b++{
				LocalMap[e].Orders[f][b] = config.OT_NoOrder
			
			}
		}

		LocalMap[e].Orders[0][config.BT_HallDown] = -1
		LocalMap[e].Orders[3][config.BT_HallUp] = -1
	}
	//LocalMap[config.My_ID].CurrentFloor = elevio.GetFloor()
	LocalMap[config.My_ID].Connected = true
	PrintMap(LocalMap)
}


func GetLocalMap() config.ElevStateMap{
	mutex.Lock()
	currentMap := LocalMap
	mutex.Unlock()
	return currentMap
}


func SetLocalMap(changedMap config.ElevStateMap){
	mutex.Lock()
	LocalMap = changedMap
	mutex.Unlock()
}


func UpdateLocalMap(changedMap config.ElevStateMap){
	currentMap := GetLocalMap()
	floorWithOpenDoor := -1

	currentMap[config.My_ID].CurrentFloor = changedMap[config.My_ID].CurrentFloor
	currentMap[config.My_ID].CurrentDir = changedMap[config.My_ID].CurrentDir
	currentMap[config.My_ID].Connected = changedMap[config.My_ID].Connected
	currentMap[config.My_ID].Door = changedMap[config.My_ID].Door

	for e:= 0; e < config.NUM_ELEVS; e++{
		if changedMap[e].Door == true && changedMap[e].Connected == true{
			floorWithOpenDoor = changedMap[e].CurrentFloor
		}
	}
	for e:= 0; e < config.NUM_ELEVS; e++{
		if currentMap[e].Connected == true { 
			for f:= 0; f < config.NUM_FLOORS; f++{
				currentMap[config.My_ID].Orders[f][config.BT_Cab] = changedMap[config.My_ID].Orders[f][config.BT_Cab]
				for b:= config.BT_HallUp; b < config.BT_Cab; b++{

					if changedMap[e].Orders[f][b] == config.OT_OrderPlaced && currentMap[e].Orders[f][b] == config.OT_NoOrder{
							currentMap[e].Orders[f][b] = changedMap[e].Orders[f][b]
						} else if changedMap[e].Orders[f][b] == config.OT_NoOrder && currentMap[e].Orders[f][b] == config.OT_OrderPlaced && floorWithOpenDoor == f{
							currentMap[e].Orders[f][b] = changedMap[e].Orders[f][b]
						}
				}
			}
		}	
	}
	
	SetLocalMap(currentMap)
	writeToBackup()
}


func UpdateMapFromNetwork(recievedMap config.ElevStateMap, newOrderChan chan config.ButtonEvent, buttonLampChan chan config.ButtonLamp){
	buttonEvent := false
	currentMap := GetLocalMap()

	floorWithOpenDoor := -1
	//PrintMap(recievedMap)
	for e:= 0; e < config.NUM_ELEVS; e++{
		if recievedMap[e].Door == true{
			floorWithOpenDoor = recievedMap[e].CurrentFloor
		}



		if e != config.My_ID {
			currentMap[e].CurrentFloor = recievedMap[e].CurrentFloor
			currentMap[e].CurrentDir = recievedMap[e].CurrentDir
			currentMap[e].Door = recievedMap[e].Door
		}
				
	}


	for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= config.BT_HallUp; b < config.BT_Cab; b++{
					if recievedMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced && currentMap[config.My_ID].Orders[f][b] == config.OT_NoOrder{
						if buttonEvent == false{
							fmt.Printf("-------------order from network--------\n")
							newOrderChan <- config.ButtonEvent{f, b}
							
							buttonEvent = true
						}
						//fmt.Printf("Order from network, floor %v, button %v\n\n", f, b)

					} else if recievedMap[config.My_ID].Orders[f][b] ==config.OT_NoOrder && currentMap[config.My_ID].Orders[f][b] == config.OT_OrderPlaced && floorWithOpenDoor == f{
						buttonLampChan <- config.ButtonLamp{f, b, false}
						for elev := 0; elev < config.NUM_ELEVS; elev++{
							currentMap[elev].Orders[f][b] = config.OT_NoOrder
						}
					}
				}
			}
	SetLocalMap(currentMap)
}

func UpdateElevStatusFromNetwork(newStatus config.StatusMsg){
	currentMap := GetLocalMap()
	currentMap[newStatus.ID].CurrentFloor = newStatus.CurrentFloor
	currentMap[newStatus.ID].CurrentDir = newStatus.CurrentDir
	currentMap[newStatus.ID].Door = newStatus.Door
	currentMap[newStatus.ID].OutOfOrder = newStatus.OutOfOrder
	SetLocalMap(currentMap)

}

func SetConnectedElevator(ID int, value bool){
	currentMap := GetLocalMap()
	currentMap[ID].Connected = value
	SetLocalMap(currentMap)
}



func PrintMap(elevMap config.ElevStateMap){
	for e := 0; e < config.NUM_ELEVS; e++ {
		fmt.Printf("\n \n \nSTATE MAP FOR ELEV %v\n", e)
		fmt.Printf("Current floor: %v \n", elevMap[e].CurrentFloor)
		fmt.Printf("Current dir: %v \n", elevMap[e].CurrentDir)
		fmt.Printf("Connected: %v \n", elevMap[e].Connected)
		fmt.Printf("Door: %v \n", elevMap[e].Door)
		
		for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= 0; b < config.NUM_BUTTONS; b++{
				fmt.Printf("%v", elevMap[e].Orders[f][b])
				fmt.Printf("\t")
			}
			fmt.Printf("\n")
		}
	}

}