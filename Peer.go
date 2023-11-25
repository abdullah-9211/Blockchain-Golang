package main

import (
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Enum to represent the type of report being set to peer's caller function
const (
	report_type_connections          = iota
	report_type_transaction_created  = iota
	report_type_block_mined          = iota
	report_type_received_transaction = iota
	report_type_received_block       = iota
	report_type_blockchain_updated   = iota
	report_type_entire_blockchain    = iota
)

// Type ReportToMain holds the information a peer sends to its calling function
type ReportToMain struct {
	Source_Address Address
	Report_Type    int
	Report_Body    string
}

// Type PeerConfig holds the settings of a peer passed to peer_main function
type PeerConfig struct {
	Self_Address          Address
	Trailing_Zeros        int
	Is_Bootstrap          bool
	Is_Miner              bool
	Is_Transaction_Maker  bool
	Bootstrap_Address     Address
	Transaction_Per_Block int
	Max_Neighbours        int
	Die_After             int64
	Up_Channel            chan ReportToMain
	Is_Bad_Node           bool
}

// type Peer holds all the information of a single peer
type Peer struct {
	Blockchain           Blockchain
	Transactions         map[Hash]Transaction
	Block_Groups         map[Hash][]Block
	Blocks               map[Hash]int64 // {Block: receive time}
	Merkel_Trees         map[Hash]MerkelTree
	My_Address           Address
	Is_Miner             bool
	Is_Transaction_Maker bool
	Is_Bootstrap         bool
	Network_Members      map[Address]int64 // {address: last contacted}
	Neighbours           map[Address]int64 // {address: last contacted}
	Bootstrap_Address    Address
	Max_Neighbours       int
	pc                   PeerConfig
}

// function peer_main creates and simulates a peer according to the given input as configuration
func peer_main(pc PeerConfig) {

	transaction_creation_channel := make(chan Transaction)
	block_mine_channel := make(chan struct {
		Block
		MerkelTree
	})
	network_packet_channel := make(chan NetworkPacket, 1)

	init_time := time.Now().Unix()
	is_mining := false
	last_neighbour_req := int64(0)
	neighbour_req_timeout := int64(10)
	last_hello := make(map[Address]int64)
	timeout := int64(15) // network members need to communicate once every timeout seconds to stay in network
	last_block_request := int64(0)

	peer := Peer{Blockchain: create_blockchain(), My_Address: pc.Self_Address, Is_Bootstrap: pc.Is_Bootstrap, Is_Miner: pc.Is_Miner, Is_Transaction_Maker: pc.Is_Transaction_Maker, Bootstrap_Address: pc.Bootstrap_Address, Max_Neighbours: pc.Max_Neighbours, Network_Members: make(map[Address]int64), Neighbours: make(map[Address]int64), Transactions: make(map[Hash]Transaction), Block_Groups: make(map[Hash][]Block), Blocks: make(map[Hash]int64), Merkel_Trees: make(map[Hash]MerkelTree), pc: pc}

	go __listen(network_packet_channel, peer.My_Address) // start listening

	if pc.Is_Transaction_Maker {
		go __transaction_creator(transaction_creation_channel)
	}

	last_print := time.Now().Unix()
	for {
		if time.Now().Unix()-last_print > 5 {
			last_print = time.Now().Unix()

			// report current active neighbours to main
			pc.Up_Channel <- ReportToMain{
				Source_Address: peer.My_Address,
				Report_Type:    report_type_connections,
				Report_Body:    fmt.Sprintf("%v", peer.Neighbours)}
		}

		// check if any node has left network
		if peer.Is_Bootstrap {
			to_drop := make([]Address, 0, 10)
			for member, last_contact := range peer.Network_Members {
				if time.Now().Unix()-last_contact > timeout {
					to_drop = append(to_drop, member)
				} else if time.Now().Unix()-last_contact > timeout/2 {
					peer.__do_hello(&last_hello, timeout, member)
				}
			}
			for _, dropping := range to_drop {
				delete(peer.Network_Members, dropping)
			}
		}

		// check if more neighbours are required
		if len(peer.Neighbours)+2 < peer.Max_Neighbours && time.Now().Unix()-last_neighbour_req > neighbour_req_timeout {
			if peer.Is_Bootstrap {
				ip_port_list := get_map_keys(peer.Network_Members)
				peer.__try_add_neighbours(ip_port_list)
			} else {
				req_packet := NetworkPacket{Req_Type: req_type_need_ip_port_list, Req_From: peer.My_Address}
				go __send_request(req_packet, pc.Bootstrap_Address)
			}
			last_neighbour_req = time.Now().Unix()
		}

		// check if any neighbour should be removed due to no contact for some time
		for neighbour, last_contact := range peer.Neighbours {
			if time.Now().Unix()-last_contact > timeout {
				delete(peer.Neighbours, neighbour)
			} else if time.Now().Unix()-last_contact > timeout/2 {
				peer.__do_hello(&last_hello, timeout, neighbour)
			}
		}

		// check if neighbour count is too high (drops random neighbours to bring it down)
		if len(peer.Neighbours) > peer.Max_Neighbours {
			peer.__drop_random_neighbours(len(peer.Neighbours) - peer.Max_Neighbours)
		}

		// start mining if peer is miner and transaction count sufficient
		if !is_mining && peer.Is_Miner && len(peer.Transactions) >= pc.Transaction_Per_Block {
			go __mine_new_block(block_mine_channel, get_map_values(peer.Transactions)[:pc.Transaction_Per_Block], peer.Blockchain.get_last_hash(), pc.Trailing_Zeros)
			is_mining = true
		}

		// check for incoming network requests
		for {
			packet, ok := CollectChanOne[NetworkPacket](network_packet_channel)
			if !ok {
				break
			}
			peer.__handle_network_packet(&packet)
		}

		// check if any transaction has been created
		for {
			transaction, ok := CollectChanOne[Transaction](transaction_creation_channel)
			if !ok {
				break
			}

			// report that a new transaction was created
			pc.Up_Channel <- ReportToMain{
				Source_Address: peer.My_Address,
				Report_Type:    report_type_transaction_created,
				Report_Body:    fmt.Sprintf("%v", transaction.Value),
			}

			peer.Transactions[transaction.hashed()] = transaction
			peer.__propagate_transaction(transaction)
		}

		// check if a block has been mined
		for is_mining {
			pair, ok := CollectChanOne[struct {
				Block
				MerkelTree
			}](block_mine_channel)
			if !ok {
				break
			}

			// report that new block was mined
			pc.Up_Channel <- ReportToMain{
				Source_Address: peer.My_Address,
				Report_Type:    report_type_block_mined,
				Report_Body:    fmt.Sprintf("%v", pair.Block.hashed().to_string()),
			}

			if peer.__extend_blockchain([]Block{pair.Block}, []MerkelTree{pair.MerkelTree}) {
				peer.__propagate_block(pair.Block, pair.MerkelTree)
			}

			is_mining = false
		}

		peer.__evaluate_block_groups(&last_block_request) // received blocks always make a group which is then either added or not added to blockchain

		// sleep to reduce load on cpu
		time.Sleep(time.Microsecond * 100)

		if pc.Die_After != -1 && time.Now().Unix()-init_time >= pc.Die_After {
			// kill itself if the die after configuration is not -1 and the given time has passed
			fmt.Printf("Port %d left the network!\n", peer.My_Address.Port)
			return
		}

	}
}

// Peer's method __drop_random_neighbours drops given number of random neighbours from the peer's neighbour list
func (peer *Peer) __drop_random_neighbours(count int) {
	ip_port_list := get_map_keys(peer.Neighbours)
	indexes := rand.Perm(len(ip_port_list))
	for i := 0; i < count; i++ {
		delete(peer.Neighbours, ip_port_list[indexes[i]])
	}
}

// Peer's method __propagate_transaction sends the provided transaction to every neighbour of the peer
func (peer *Peer) __propagate_transaction(transaction Transaction) {
	packet_to_send := NetworkPacket{Req_Type: req_type_new_transaction, Req_From: peer.My_Address, Transaction: transaction}
	for neighbour := range peer.Neighbours {
		go __send_request(packet_to_send, neighbour)
	}
}

// Peer's method __propagate_block sends the provided block to every  neighbour of the peer
func (peer *Peer) __propagate_block(block Block, merkel_tree MerkelTree) {
	packet_to_send := NetworkPacket{Req_Type: req_type_new_block, Req_From: peer.My_Address, Block: block, Merkel_Tree: merkel_tree}
	for neighbour := range peer.Neighbours {
		go __send_request(packet_to_send, neighbour)
	}
}

// Peer's method __extend_blockchain extends blockchain with the given blocks and merkel trees.
// if the blockchain accepts the new blocks, return true else return false
func (peer *Peer) __extend_blockchain(blocks []Block, merkel_trees []MerkelTree) bool {
	if len(blocks) != len(merkel_trees) {
		return false
	}
	for i := 0; i < len(blocks); i++ {
		if blocks[i].Merkel_Root != merkel_trees[i].hashed() || !blocks[i].is_valid() || !merkel_trees[i].is_valid() { // checks that the blocks and merkel trees are valid and compatible
			return false
		}
	}
	_, already_in := peer.Blockchain.Blocks[blocks[len(blocks)-1].hashed()]
	if already_in {
		return false
	}
	all_transactions := make(map[Hash]bool)
	for i := 0; i < len(blocks); i++ {
		peer.Blockchain.add_block(blocks[i])
		peer.Blockchain.add_merkel_tree(merkel_trees[i])
		for transaction_hash := range merkel_trees[i].Transactions {
			all_transactions[transaction_hash] = true
		}
	}
	peer.Blockchain.remove_short_chains()
	_, added := peer.Blockchain.Blocks[blocks[len(blocks)-1].hashed()]
	if !added {
		return false
	}
	pruned_transactions := make(map[Hash]Transaction)
	for _, transaction := range peer.Transactions {
		_, ok := all_transactions[transaction.hashed()]
		if !ok {
			pruned_transactions[transaction.hashed()] = transaction
		}
	}
	peer.Transactions = pruned_transactions

	// report a change in blockchain to main
	peer.pc.Up_Channel <- ReportToMain{
		Source_Address: peer.My_Address,
		Report_Type:    report_type_blockchain_updated,
		Report_Body:    "",
	}

	// report the new state of blockchain to main
	peer.pc.Up_Channel <- ReportToMain{
		Source_Address: peer.My_Address,
		Report_Type:    report_type_entire_blockchain,
		Report_Body:    peer.Blockchain.pretty_print(),
	}

	return true
}

// Peer's method __do_hello sends a hello to a neighbour if the time since the last hello was sent is >= timeout / 8
func (peer *Peer) __do_hello(last_hello *map[Address]int64, timeout int64, target Address) {
	last_hello_time, ok := (*last_hello)[target]
	if !ok {
		last_hello_time = 0
	}
	if time.Now().Unix()-last_hello_time < timeout/8 {
		return
	}
	packet_to_send := NetworkPacket{Req_Type: req_type_hello, Req_From: peer.My_Address}
	go __send_request(packet_to_send, target)
	(*last_hello)[target] = time.Now().Unix()
}

// Peer's method __evaluate_block_groups
func (peer *Peer) __evaluate_block_groups(last_block_request *int64) {

	// block groups to remove
	to_remove := make([]Hash, 0, len(peer.Block_Groups))

	for prev_hash, blocks := range peer.Block_Groups {

		// remove group if time since a block was added to it exceeds 60
		if time.Now().Unix()-peer.Blocks[prev_hash] > 60 {
			to_remove = append(to_remove, prev_hash)
			continue
		}

		// checks whether the block group has a previous block in the blockchain and can thus, be added to the chain
		_, in_chain := peer.Blockchain.Blocks[prev_hash]
		if prev_hash == (Hash{}) {
			in_chain = true
		}

		if in_chain {

			// try adding block group to chain
			blocks = reverse_slice(blocks)
			merkel_trees := make([]MerkelTree, 0, len(blocks))
			for _, block := range blocks {
				merkel_trees = append(merkel_trees, peer.Merkel_Trees[block.Merkel_Root])
			}
			if peer.__extend_blockchain(blocks, merkel_trees) {
				// propagate the block if the new block has made a change to the blockchain
				peer.__propagate_block(blocks[len(blocks)-1], merkel_trees[len(merkel_trees)-1])
			}

			// remove the block group that was potentially added to chain
			to_remove = append(to_remove, prev_hash)

		} else if time.Now().Unix()-*last_block_request > 0 {

			// request the neighbours for the block that should come before the earliest block in the given block group
			packet_to_send := NetworkPacket{Req_Type: req_type_need_block, Req_From: peer.My_Address, Block_Hash: prev_hash}
			for neighbour := range peer.Neighbours {
				go __send_request(packet_to_send, neighbour)
			}
			*last_block_request = time.Now().Unix()
		}
	}

	// remove block groups that are required to be removed
	for _, prev_hash := range to_remove {
		delete(peer.Block_Groups, prev_hash)
		delete(peer.Blocks, prev_hash)
	}
}

// Peer's method __try_add_neighbours sends neighbour connection request to random peers in the given list
func (peer *Peer) __try_add_neighbours(ip_port_list []Address) {
	need_neighbours := max(0, peer.Max_Neighbours-len(peer.Neighbours)-2) // two reserved for anyone who wants to connect to this peer
	indexes := rand.Perm(len(ip_port_list))
	for _, index := range indexes {
		if need_neighbours == 0 {
			break
		}

		target_peer := ip_port_list[index]
		_, already_neighbour := peer.Neighbours[target_peer]
		if target_peer == peer.My_Address || already_neighbour {
			continue
		}

		packet_to_send := NetworkPacket{Req_Type: req_type_new_connection, Req_From: peer.My_Address}
		go __send_request(packet_to_send, target_peer)
		need_neighbours--
	}
}

// Peer's method __handle_network_packet deals with the given packet as per requirement
func (peer *Peer) __handle_network_packet(packet *NetworkPacket) {
	_, in_neighbours := peer.Neighbours[packet.Req_From]
	switch packet.Req_Type {
	case req_type_new_connection:
		if len(peer.Neighbours) < peer.Max_Neighbours {
			peer.Neighbours[packet.Req_From] = time.Now().Unix()
			go __send_request(NetworkPacket{Req_Type: req_type_accept_connection, Req_From: peer.My_Address}, packet.Req_From)
		} else {
			go __send_request(NetworkPacket{Req_Type: req_type_reject_connection, Req_From: peer.My_Address}, packet.Req_From)
		}
	case req_type_accept_connection:
		peer.Neighbours[packet.Req_From] = time.Now().Unix()
	case req_type_new_transaction:
		if !in_neighbours {
			return
		}
		_, already_exists := peer.Transactions[packet.Transaction.hashed()]
		if !already_exists {
			peer.Transactions[packet.Transaction.hashed()] = packet.Transaction // add transaction to list of transactions
			packet_to_send := NetworkPacket{Req_Type: req_type_new_transaction, Req_From: peer.My_Address, Transaction: packet.Transaction}
			for neighbour := range peer.Neighbours {
				if neighbour != packet.Req_From {
					go __send_request(packet_to_send, neighbour) // propagate transaction to all neighbours except sender
				}
			}

			// report that a new transaction has been received
			peer.pc.Up_Channel <- ReportToMain{
				Source_Address: peer.My_Address,
				Report_Type:    report_type_received_transaction,
				Report_Body:    fmt.Sprintf("%v from %v", packet.Transaction.Value, packet.Req_From),
			}

		}
	case req_type_new_block:
		if !in_neighbours {
			return
		}
		if peer.pc.Is_Bad_Node || packet.Block.Trailing_Zeros < peer.pc.Trailing_Zeros || !packet.Block.is_valid() || !packet.Merkel_Tree.is_valid() {
			return
		}
		block_hash := packet.Block.hashed()
		prev_data, in_groups := peer.Block_Groups[block_hash]
		if in_groups {
			delete(peer.Block_Groups, block_hash)
			delete(peer.Blocks, block_hash)
			prev_data = append(prev_data, packet.Block)
		} else {
			prev_data = []Block{packet.Block}
		}
		peer.Merkel_Trees[packet.Merkel_Tree.hashed()] = packet.Merkel_Tree
		peer.Block_Groups[packet.Block.Prev_Block] = prev_data
		peer.Blocks[packet.Block.Prev_Block] = time.Now().Unix()

		// report that a new block has been received
		peer.pc.Up_Channel <- ReportToMain{
			Source_Address: peer.My_Address,
			Report_Type:    report_type_received_block,
			Report_Body:    fmt.Sprintf("%v from %v", packet.Block.hashed().to_string(), packet.Req_From),
		}
	case req_type_need_block:
		block, exists := peer.Blockchain.Blocks[packet.Block_Hash]
		if exists {
			packet_to_send := NetworkPacket{Req_Type: req_type_new_block, Req_From: peer.My_Address, Block: block, Merkel_Tree: peer.Merkel_Trees[block.Merkel_Root]}
			go __send_request(packet_to_send, packet.Req_From)
		}
	case req_type_need_ip_port_list:
		if peer.Is_Bootstrap {
			network_members := get_map_keys(peer.Network_Members)
			new_packet := NetworkPacket{Req_Type: req_type_ip_port_list, Req_From: peer.My_Address, Ip_Port_List: network_members}
			peer.Network_Members[packet.Req_From] = time.Now().Unix()
			go __send_request(new_packet, packet.Req_From)
		}
	case req_type_ip_port_list:
		if packet.Req_From == peer.Bootstrap_Address {
			peer.__try_add_neighbours(packet.Ip_Port_List)
		}
	case req_type_hello:
		if in_neighbours || packet.Req_From == peer.Bootstrap_Address || peer.Is_Bootstrap {
			if in_neighbours {
				peer.Neighbours[packet.Req_From] = time.Now().Unix()
			}
			if peer.Is_Bootstrap {
				peer.Network_Members[packet.Req_From] = time.Now().Unix()
			}
			go __send_request(NetworkPacket{Req_Type: req_type_hi, Req_From: peer.My_Address}, packet.Req_From)
		}
	case req_type_hi:
		if in_neighbours || packet.Req_From == peer.Bootstrap_Address || peer.Is_Bootstrap {
			if in_neighbours {
				peer.Neighbours[packet.Req_From] = time.Now().Unix()
			}
			if peer.Is_Bootstrap {
				peer.Network_Members[packet.Req_From] = time.Now().Unix()
			}
		}
	}
}

// function __transaction_creator runs infinitely and creates a random transaction after a random interval and
// sends it to the channel that was passed to this function as a input
func __transaction_creator(up_channel chan<- Transaction) {
	var second int64 = 1000000000
	for {
		time.Sleep(time.Duration(random_int(2*second, 8*second)))
		transaction := Transaction{Value: random_string(15), Not_Null: true}
		up_channel <- transaction
	}
}

// function __mine_new_block uses the given transactions, prev_hash value, and trailing_zeros count to create a new block
// and mine a valid nonce value. once complete, the block is written to the channel that was passed to this function as input
func __mine_new_block(up_channel chan<- struct {
	Block
	MerkelTree
}, transactions []Transaction, prev_hash Hash, trailing_zeros int) {
	merkel_tree := create_merkel_tree()
	for _, transaction := range transactions {
		merkel_tree.add_transaction(transaction)
	}
	merkel_tree.build()
	block := create_block(merkel_tree.hashed(), prev_hash, trailing_zeros)
	block.mine()
	up_channel <- struct {
		Block
		MerkelTree
	}{block, merkel_tree}
}

// function __listen listens a given address and writes any packet it recieves to the channel that was passed to it as input
func __listen(up_channel chan<- NetworkPacket, address Address) {
	ln, err := net.Listen("tcp", address.to_string())
	if err != nil {
		fmt.Println("Error starting listening", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Network Listen at %d failed\n", address.Port)
			continue
		}
		go __receive_request(up_channel, conn)
	}
}

// function send request sends the given network packet to the given target address from a random port
func __send_request(network_packet NetworkPacket, target Address) {
	conn, err := net.Dial("tcp", target.to_string())
	if err != nil {
		fmt.Printf("Network Dial from %d to %d failed: %v\n", network_packet.Req_From.Port, target.Port, err)
		return
	}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(&network_packet)
	conn.Close()
}

// function receive request uses a connection and receives the network packet sent on it. this is then written to
// channel that is passed to it as input
func __receive_request(up_channel chan<- NetworkPacket, conn net.Conn) {
	dec := gob.NewDecoder(conn)
	network_packet := &NetworkPacket{}
	dec.Decode(&network_packet)
	up_channel <- *network_packet
	conn.Close()
}
