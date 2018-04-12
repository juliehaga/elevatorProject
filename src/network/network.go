package network

import (
	"../elevStateMap"
	"../config"
	"./bcast"
	"fmt"
	"net"
	"time"
	"encoding/json"
)


const interval = 15 * time.Millisecond
const timeout = 1500 * time.Millisecond





func Network(messageRx chan config.Message, messageTx chan config.Message, statusMsgRx chan config.StatusMsg, orderMsgRx chan config.OrderMsg, activeOrderRx chan config.ActiveOrders){

	go bcast.Transmitter(16666, messageTx)
	go bcast.Receiver(16666, messageRx)


	for {
		select{
			case receivedMsg := <-messageRx:

				if receivedMsg.MsgType == config.ElevStatus{
					statusMsgRx <- config.StatusMsg{receivedMsg.ID, receivedMsg.ElevMap[receivedMsg.ID].CurrentFloor, receivedMsg.ElevMap[receivedMsg.ID].CurrentDir, receivedMsg.ElevMap[receivedMsg.ID].Door, receivedMsg.ElevMap[receivedMsg.ID].OutOfOrder,receivedMsg.ElevMap[receivedMsg.ID].IDLE}

				} else if receivedMsg.MsgType == config.Orders {
					orderMsgRx <- config.OrderMsg{receivedMsg.ID, receivedMsg.ElevMap}

				} else if receivedMsg.MsgType == config.ActiveOrder{
					fmt.Printf("Mottar en ordremsg fra %v\n", receivedMsg.ID)
					activeOrderRx <- config.ActiveOrders{receivedMsg.Button, receivedMsg.ID, true}
				}

		}
	}
	
}


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
































