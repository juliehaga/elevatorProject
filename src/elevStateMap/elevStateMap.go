package elevStateMap

import(
	"../config"
	"../elevio"	
	"fmt"
	"sync"
	"os"
	"strconv"
	"io"
)


var LocalMap ElevStateMap
var mutex = &sync.Mutex{}


type OrderType int
const (
	OT_NoOrder                = 0
	OT_OrderPlaced            = 1
)

type ElevDir int

const (
	ED_Up   		ElevDir = 1
	ED_Down                 = -1
)



type ElevInfo struct{
	CurrentFloor int
	CurrentDir ElevDir
	Connected bool
	Door bool
	IDLE bool
	Orders [config.NUM_FLOORS][config.NUM_BUTTONS] OrderType
}


type ElevStateMap [config.NUM_ELEVS]ElevInfo

type Message struct {
	ID int
	MsgType int
	ElevMap   ElevStateMap
}


type StatusMsg struct {
	ID int
	CurrentFloor int
	CurrentDir ElevDir
	Door bool
}

type OrderMsg struct{
	ID int
	ElevMap ElevStateMap
}


func writeToBackup(){
	// open file using READ & WRITE permission
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()

	// write some text line-by-line to file

	for f := 0; f< config.NUM_FLOORS; f++{
		var order int = int(LocalMap[config.My_ID].Orders[f][elevio.BT_Cab])
		_, err = file.WriteString(strconv.Itoa(order) + "\n")
		if err != nil { return }
	}
	// save changes
	err = file.Sync()
	if err != nil { return }

	fmt.Println("==> done writing to file")
}

func readFromBackup(){
	// re-open file
	var file, err = os.OpenFile("backup.txt", os.O_RDWR, 0644)
	if err != nil { return }
	defer file.Close()
	currentMap := GetLocalMap()

	// read file, line by line
	var buf = make([]byte, 1024)
	floor := 0
	for {

		_, err = file.Read(buf)
		order, _ := strconv.Atoi(string(buf))


		currentMap[config.My_ID].Orders[floor][elevio.BT_Cab] = OrderType(order)

		floor++
		
		// break if finally arrived at end of file
		if err == io.EOF {
			break
		}
		
		// break if error occured
		if err != nil && err != io.EOF {
			break
		}
	}
	SetLocalMap(currentMap)
	
}

func InitElevStateMap(){
	var _, err = os.Stat("backup.txt")

	if os.IsNotExist(err) {
		fmt.Printf("Creating new backup file\n")
		var file, err = os.Create("backup.txt")
		if err != nil{
			fmt.Printf("Could not create file \n")
			//return
		}	
		for f := 0; f < config.NUM_FLOORS; f++{
			LocalMap[config.My_ID].Orders[f][elevio.BT_Cab] = OT_NoOrder
		}
		writeToBackup()
		defer file.Close()
	}else{
		readFromBackup()
	}


	//initialize map	
	for e:= 0; e < config.NUM_ELEVS; e++{
		LocalMap[e].IDLE = true
		LocalMap[e].CurrentDir = ED_Down
		LocalMap[e].Connected = false
		LocalMap[e].Door = false
		LocalMap[e].CurrentFloor = -1
			
		for f := 0; f < config.NUM_FLOORS; f++{
			for b :=0; b < config.NUM_BUTTONS; b++{
				LocalMap[e].Orders[f][b] = OT_NoOrder
			
			}
		}

		LocalMap[e].Orders[0][elevio.BT_HallDown] = -1
		LocalMap[e].Orders[3][elevio.BT_HallUp] = -1
	}
	LocalMap[config.My_ID].CurrentFloor = elevio.GetFloor()
	LocalMap[config.My_ID].Connected = true
}


func GetLocalMap() ElevStateMap{
	mutex.Lock()
	currentMap := LocalMap
	mutex.Unlock()
	return currentMap
}


func SetLocalMap(changedMap ElevStateMap){
	mutex.Lock()
	LocalMap = changedMap
	mutex.Unlock()
}


func UpdateLocalMap(changedMap ElevStateMap){
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
				currentMap[config.My_ID].Orders[f][elevio.BT_Cab] = changedMap[config.My_ID].Orders[f][elevio.BT_Cab]
				for b:= elevio.BT_HallUp; b < elevio.BT_Cab; b++{

					if changedMap[e].Orders[f][b] == OT_OrderPlaced && currentMap[e].Orders[f][b] == OT_NoOrder{
							currentMap[e].Orders[f][b] = changedMap[e].Orders[f][b]
						} else if changedMap[e].Orders[f][b] == OT_NoOrder && currentMap[e].Orders[f][b] == OT_OrderPlaced && floorWithOpenDoor == f{
							currentMap[e].Orders[f][b] = changedMap[e].Orders[f][b]
						}
				}
			}
		}	
	}
	//fmt.Printf("*******************MITT MAP etter endring******************")

	//PrintMap(currentMap)
	SetLocalMap(currentMap)


}


func UpdateMapFromNetwork(recievedMap ElevStateMap, newOrderChan chan elevio.ButtonEvent, buttonLampChan chan elevio.ButtonLamp){
	currentMap := GetLocalMap()

	floorWithOpenDoor := -1
	//PrintMap(recievedMap)
	for e:= 0; e < config.NUM_ELEVS; e++{
		if currentMap[e].Connected == true{
			if recievedMap[e].Door == true{
				floorWithOpenDoor = recievedMap[e].CurrentFloor
			}

			//sjekk om heis e er i live
			if recievedMap[e].Connected == true{

				if e != config.My_ID {
					currentMap[e].CurrentFloor = recievedMap[e].CurrentFloor
					currentMap[e].CurrentDir = recievedMap[e].CurrentDir
					currentMap[e].Door = recievedMap[e].Door
				}
				
			}
		}
	}


	for f:= 0; f < config.NUM_FLOORS; f++{
			for b:= elevio.BT_HallUp; b < elevio.BT_Cab; b++{
					if recievedMap[config.My_ID].Orders[f][b] == OT_OrderPlaced && currentMap[config.My_ID].Orders[f][b] == OT_NoOrder{
						

						newOrderChan <- elevio.ButtonEvent{f, b}
						
						//fmt.Printf("Order from network, floor %v, button %v\n\n", f, b)

					} else if recievedMap[config.My_ID].Orders[f][b] == OT_NoOrder && currentMap[config.My_ID].Orders[f][b] == OT_OrderPlaced && floorWithOpenDoor == f{
						//fmt.Printf("Ordered completed from netowrk floor %v, button %v\n", f, b)
						//clear orders from all elevators
						buttonLampChan <- elevio.ButtonLamp{f, b, false}
						for elev := 0; elev < config.NUM_ELEVS; elev++{

							currentMap[elev].Orders[f][b] = OT_NoOrder
						}
					}
				}
			}
	SetLocalMap(currentMap)
}

func UpdateElevStatusFromNetwork(newStatus StatusMsg){
	currentMap := GetLocalMap()
	currentMap[newStatus.ID].CurrentFloor = newStatus.CurrentFloor
	currentMap[newStatus.ID].CurrentDir = newStatus.CurrentDir
	currentMap[newStatus.ID].Door = newStatus.Door
	SetLocalMap(currentMap)

}

func SetConnectedElevator(ID int, value bool){
	currentMap := GetLocalMap()
	currentMap[ID].Connected = value
	SetLocalMap(currentMap)
}




func PrintMap(elevMap ElevStateMap){
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