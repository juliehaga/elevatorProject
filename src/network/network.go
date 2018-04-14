package network

import (
	"../elevStateMap"
	"../config"
	"./bcast"
	"fmt"
	"time"
)


const interval = 15 * time.Millisecond
const timeout = 1500 * time.Millisecond





func Network(messageRx chan config.Message, messageTx chan config.Message, statusMsgRx chan config.StatusMsg, orderMsgRx chan config.OrderMsg, activeOrderRx chan config.ActiveOrders,  ackChan chan config.AckMsg){

	go bcast.Transmitter(16666, messageTx)
	go bcast.Receiver(16666, messageRx)

	for {
		select{
			case receivedMsg := <-messageRx:
				if receivedMsg.ID != config.My_ID{

					if receivedMsg.MsgType == config.ElevStatus{
						statusMsgRx <- config.StatusMsg{receivedMsg.ID, receivedMsg.ElevMap[receivedMsg.ID].CurrentFloor, receivedMsg.ElevMap[receivedMsg.ID].CurrentDir, receivedMsg.ElevMap[receivedMsg.ID].Door, receivedMsg.ElevMap[receivedMsg.ID].Connected}
						SendAck(messageTx, receivedMsg.ElevMap, receivedMsg.ID)
					} else if receivedMsg.MsgType == config.Orders {
						orderMsgRx <- config.OrderMsg{receivedMsg.ID, receivedMsg.ElevMap}
						SendAck(messageTx, receivedMsg.ElevMap, receivedMsg.ID)
					} else if receivedMsg.MsgType == config.ActiveOrder{
						fmt.Printf("Mottar en ordremsg fra %v , om knapp: %v \n", receivedMsg.ID, receivedMsg.Button)
						activeOrderRx <- config.ActiveOrders{receivedMsg.Button, receivedMsg.ID, true, receivedMsg.ElevMap}
						SendAck(messageTx, receivedMsg.ElevMap, receivedMsg.ID)
					} else if receivedMsg.MsgType == config.Ack{
						fmt.Printf("Mottar ack msg\n")
						ackChan <- config.AckMsg{receivedMsg.ID, receivedMsg.Reciever_ID}
					}
				}

		}
	}
	
}


func SendOrders(messageTx chan config.Message, elevMap config.ElevStateMap, ackChan chan config.AckMsg) {
	elevMapMsg := config.Message{config.My_ID, config.Orders, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	waitForAck(elevMapMsg, messageTx, ackChan)
}

func SendElevStatus(messageTx chan config.Message,  elevMap config.ElevStateMap, ackChan chan config.AckMsg){
	elevMapMsg := config.Message{config.My_ID, config.ElevStatus, elevMap, -1, config.ButtonEvent{0, config.BT_HallDown}}
	waitForAck(elevMapMsg, messageTx, ackChan)
}


func SendActiveOrder(messageTx chan config.Message,  order config.ActiveOrders, ackChan chan config.AckMsg){
	elevMap := elevStateMap.GetLocalMap()
	elevMapMsg := config.Message{config.My_ID, config.ActiveOrder, elevMap, -1, order.Button}
	waitForAck(elevMapMsg, messageTx,ackChan)
}
func SendAck(messageTx chan config.Message,  elevMap config.ElevStateMap, recieverID int){
	fmt.Printf("Jeg sender ACK\n")
	AckMsg := config.Message{config.My_ID, config.Ack, elevMap, recieverID, config.ButtonEvent{0, config.BT_HallDown}}
	messageTx <- AckMsg
}

func waitForAck(message config.Message, messageTx chan config.Message, ackChan chan config.AckMsg){
	for e:= 0; e < config.NUM_ELEVS; e++{
		if e != config.My_ID{
			message.Reciever_ID = e
					
			if message.ElevMap[e].Connected == true{
							
				WAIT_FOR_ACK:
					for i := 0; i < 5; i++{
						messageTx <- message
						select {
							case ackMsg := <- ackChan:
								if ackMsg.Reciever_ID == config.My_ID && ackMsg.Transmitter_ID == e{
									break WAIT_FOR_ACK
									fmt.Printf("Avbryter, fikk ACK%v\n", ackMsg.Transmitter_ID)
								}
							default:
						}
						fmt.Printf("legger på transmit channel\n")
						time.Sleep(10* time.Millisecond)
					}
			}
		}
	}
}
































