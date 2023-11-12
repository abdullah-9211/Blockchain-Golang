package main

import (
	"fmt"
)

// constants for request type ids
const (
	req_type_new_connection    = iota
	req_type_accept_connection = iota
	req_type_reject_connection = iota
	req_type_new_transaction   = iota
	req_type_new_block         = iota
	req_type_need_block        = iota
	req_type_need_ip_port_list = iota
	req_type_ip_port_list      = iota
	req_type_hello             = iota
	req_type_hi                = iota
)

type Address struct {
	Ip   uint32
	Port uint16
}

func (address Address) to_string() string {
	if address.Ip == 0 {
		return fmt.Sprintf("localhost:%d", address.Port)
	}
	return fmt.Sprintf("%d:%d", address.Ip, address.Port)
}

type NetworkPacket struct {
	Req_Type     int
	Req_From     Address // ip and port number
	Transaction  Transaction
	Block        Block
	Merkel_Tree  MerkelTree
	Ip_Port_List []Address
	Block_Hash   Hash
}

func create_network_packet(req_type int, req_from Address) NetworkPacket {
	return NetworkPacket{Req_Type: req_type, Req_From: req_from}
}
