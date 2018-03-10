package network

import (
	"../elevStateMap"
	"../config"
	"fmt"
	"net"
	"sort"
	"time"
	"encoding/json"
	"reflect"
	"os"
	"syscall"
	"strconv"

)







const interval = 15 * time.Millisecond
const timeout = 1500 * time.Millisecond






func SendOrders(messageTx chan config.Message, elevMap config.ElevStateMap) {
	elevMapMsg := config.Message{config.My_ID, config.Orders, elevMap}
	messageTx <- elevMapMsg
}

func SendElevStatus(messageTx chan config.Message,  elevMap config.ElevStateMap){
	elevMapMsg := config.Message{config.My_ID, config.ElevStatus, elevMap}
	messageTx <- elevMapMsg
}

func SendAck(messageTx chan config.Message,  elevMap config.ElevStateMap){
	AckMsg := config.Message{config.My_ID, config.Ack, elevMap}
	messageTx <- AckMsg

}

func PeerTransmitter(port int, id string, transmitEnable <-chan bool) {

	conn := DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func PeerReceiver(port int, peerUpdateCh chan<- config.PeerUpdate) {

	var buf [1024]byte
	var p config.PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}
		//Kan her sette connected lik false?
		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)

			for i:= 0; i < len(p.Lost); i ++{
				lostID, _ := strconv.Atoi(p.Lost[i])
				elevStateMap.SetConnectedElevator(lostID, false)

			}
			newID, _ := strconv.Atoi(p.New)
			fmt.Printf("Setter til connected %v", newID)
			

			elevStateMap.SetConnectedElevator(newID, true)

			peerUpdateCh <- p
		}
	}
}


func Transmitter(port int, messageTx chan config.Message, ackChan chan config.AckMsg){
	for {
		select {
			case message := <- messageTx:

				addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", port))
				conn, _ := net.DialUDP("udp", nil, addr)
				buf, _ := json.Marshal(message)


				for e:= 0; e < config.NUM_ELEVS; e++{
					if e != config.My_ID{
						if message.ElevMap[e].Connected == true{
							for i := 0; i < 5; i++{
								conn.Write(buf)
								select {
									case ackMsg := <- ackChan:
										if ackMsg.ID == e {
											fmt.Printf("Recieved ack from %v\n", ackMsg.ID)
											break
										}
								}
								//antar at peer vil fiksa å sette til dead dersom en faller ut.
							}
						}
					}
				}
		}
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(port int, orderMsgRx chan config.OrderMsg, statusMsgRx chan config.StatusMsg, ackChan chan config.AckMsg, messageTx chan config.Message) {
	//var receivedMap elevStateMap.ElevStateMap
	var receivedMsg config.Message
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", port))
	conn, _ := net.ListenUDP("udp", addr)

	var b[1048576] byte	
	elevMap := elevStateMap.GetLocalMap()
	for {
		integer, _, err := conn.ReadFromUDP(b[:])
		if err != nil {
			fmt.Printf("Feil %s", err. Error())
		}
		if integer > 0 {

			json.Unmarshal(b[:integer], &receivedMsg)
			if receivedMsg.MsgType == config.ElevStatus{	
				statusMsgRx <- config.StatusMsg{receivedMsg.ID, receivedMsg.ElevMap[receivedMsg.ID].CurrentFloor, receivedMsg.ElevMap[receivedMsg.ID].CurrentDir, receivedMsg.ElevMap[receivedMsg.ID].Door, receivedMsg.ElevMap[receivedMsg.ID].OutOfOrder}
				SendAck(messageTx, elevMap)
			} else if receivedMsg.MsgType == config.Orders {
				orderMsgRx <- config.OrderMsg{receivedMsg.ID, receivedMsg.ElevMap}
				SendAck(messageTx, elevMap)
			} else if receivedMsg.MsgType == config.Ack{
				ackChan <- config.AckMsg{receivedMsg.ID}
			}
			
		} else {
			conn.Close()
		}
	}
	
}



// Checks that args to Tx'er/Rx'er are valid:
//  All args must be channels
//  Element types of channels must be encodable with JSON
//  No element types are repeated
// Implementation note:
//  - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//    so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg#%d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg#%d and arg#%d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		switch elemType.Kind() {
		case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (arg#%d)",
				elemType.String(), i+1))
		case reflect.Map:
			if elemType.Key().Kind() != reflect.String {
				panic(fmt.Sprintf(
					"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (arg#%d)",
					elemType.String(), i+1))
			}
		}
	}
}


func DialBroadcastUDP(port int) net.PacketConn {
	s, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})

	f := os.NewFile(uintptr(s), "")
	conn, _ := net.FilePacketConn(f)
	f.Close()

	return conn
}