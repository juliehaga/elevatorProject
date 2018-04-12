package network

import (
	"../elevStateMap"
	"../config"
	"fmt"
	"net"
	"time"
	"encoding/json"
)


const interval = 15 * time.Millisecond
const timeout = 1500 * time.Millisecond


func SendOrders(messageTx chan config.Message, elevMap config.ElevStateMap) {
	elevMapMsg := config.Message{config.My_ID, config.Orders, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	messageTx <- elevMapMsg
}

func SendElevStatus(messageTx chan config.Message,  elevMap config.ElevStateMap){
	elevMapMsg := config.Message{config.My_ID, config.ElevStatus, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	messageTx <- elevMapMsg
}

func SendAck(messageTx chan config.Message,  elevMap config.ElevStateMap, recieverID int, port int){
	AckMsg := config.Message{config.My_ID, config.Ack, elevMap, recieverID, config.ButtonEvent{0, config.BT_HallDown}}
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", port))
	conn, _ := net.DialUDP("udp", nil, addr)
	buf, _ := json.Marshal(AckMsg)		
	conn.Write(buf)
}

func SendActiveOrder(messageTx chan config.Message,  order config.ActiveOrders){
	elevMap := elevStateMap.GetLocalMap()
	elevMapMsg := config.Message{config.My_ID, config.ActiveOrder, elevMap, -1, order.Button}
	messageTx <- elevMapMsg
}





























