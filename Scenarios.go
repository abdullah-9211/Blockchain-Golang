package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"time"
)

func handle_reports(report chan ReportToMain, to_print []int) {
	set := uint64(0)
	for _, value := range to_print {
		set |= (1 << value)
	}

	connections := make(map[string]string)

	for {
		time.Sleep(time.Millisecond) // added to prevent too high cpu usage
		report, ok := CollectChanOne(report)
		if !ok {
			continue
		}

		if report.Report_Type == report_type_connections && bit_is_set(set, report_type_connections) {
			if connections[report.Source_Address.to_string()] != report.Report_Body {
				connections[report.Source_Address.to_string()] = report.Report_Body
				tmp := make([]string, 0, len(connections))
				for src, body := range connections {
					tmp = append(tmp, fmt.Sprintf("%v - Connections: %v\n", src, body))
				}
				sort.StringSlice.Sort(tmp)
				to_write := ""
				for _, ele := range tmp {
					to_write += ele
				}
				write_to_file("connections.txt", to_write)
			}
		}

		if report.Report_Type == report_type_transaction_created && bit_is_set(set, report_type_transaction_created) {
			fmt.Printf(
				"%v - Created a Transaction: %v\n",
				report.Source_Address.to_string(),
				report.Report_Body)
		}

		if report.Report_Type == report_type_block_mined && bit_is_set(set, report_type_block_mined) {
			fmt.Printf(
				"%v - Mined a Block: %v\n",
				report.Source_Address.to_string(),
				report.Report_Body)
		}

		if report.Report_Type == report_type_received_transaction && bit_is_set(set, report_type_received_transaction) {
			fmt.Printf(
				"%v - Received a Transaction: %v\n",
				report.Source_Address.to_string(),
				report.Report_Body)
		}

		if report.Report_Type == report_type_received_block && bit_is_set(set, report_type_received_block) {
			fmt.Printf(
				"%v - Received a Block: %v\n",
				report.Source_Address.to_string(),
				report.Report_Body)
		}

		if report.Report_Type == report_type_blockchain_updated && bit_is_set(set, report_type_blockchain_updated) {
			fmt.Printf(
				"%v - Updated its blockchain\n",
				report.Source_Address.to_string())
		}

		if report.Report_Type == report_type_entire_blockchain && bit_is_set(set, report_type_entire_blockchain) {
			filename := fmt.Sprintf("Blockchain_%d.txt", report.Source_Address.Port)
			write_to_file(filename, report.Report_Body)
		}

	}
}

func scenario_connection_control() {

	reports := make(chan ReportToMain, 100)
	go handle_reports(reports, []int{report_type_connections})

	max_neighbours := 3
	bootstrap_address := Address{Port: 8080}

	peer_config := PeerConfig{
		Self_Address:          Address{Port: 8080},
		Trailing_Zeros:        20,
		Is_Bootstrap:          true,
		Is_Miner:              false,
		Is_Transaction_Maker:  false,
		Bootstrap_Address:     Address{},
		Transaction_Per_Block: 4,
		Max_Neighbours:        max_neighbours,
		Die_After:             int64(-1),
		Up_Channel:            reports,
	}

	go peer_main(peer_config)

	peer_config.Bootstrap_Address = bootstrap_address
	peer_config.Is_Bootstrap = false

	for i := 8081; i <= 8088; i++ {

		if i == 8081 {
			peer_config.Die_After = 30
		}

		peer_config.Self_Address = Address{Port: uint16(i)}
		go peer_main(peer_config)

		if i == 8081 {
			peer_config.Die_After = -1
		}

	}

	time.Sleep(30 * time.Second)

	peer_config.Self_Address = Address{Port: 8089}
	go peer_main(peer_config)
	fmt.Printf("Port %d joined the network!\n", 8089)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func scenario_blockchain_observation() {

	reports := make(chan ReportToMain, 100)
	go handle_reports(reports, []int{report_type_received_transaction,
		report_type_transaction_created,
		report_type_block_mined,
		report_type_received_block,
		report_type_blockchain_updated,
		report_type_entire_blockchain,
		report_type_connections})

	max_neighbours := 3
	bootstrap_address := Address{Port: 8080}

	peer_config := PeerConfig{
		Self_Address:          Address{Port: 8080},
		Trailing_Zeros:        20,
		Is_Bootstrap:          true,
		Is_Miner:              false,
		Is_Transaction_Maker:  false,
		Bootstrap_Address:     Address{},
		Transaction_Per_Block: 4,
		Max_Neighbours:        max_neighbours,
		Die_After:             int64(-1),
		Up_Channel:            reports,
	}

	go peer_main(peer_config)

	peer_config.Bootstrap_Address = bootstrap_address
	peer_config.Is_Bootstrap = false

	for i := 8081; i <= 8086; i++ {
		peer_config.Is_Miner = i%2 == 1
		peer_config.Is_Transaction_Maker = i%2 == 0
		peer_config.Self_Address = Address{Port: uint16(i)}

		go peer_main(peer_config)
	}

	time.Sleep(30 * time.Second)

	peer_config.Is_Miner = true
	peer_config.Is_Transaction_Maker = true
	peer_config.Self_Address = Address{Port: 8089}
	go peer_main(peer_config)
	fmt.Printf("Port %d joined the network!\n", 8089)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func scenario_blockchain_bad_node() {

	reports := make(chan ReportToMain, 100)
	go handle_reports(reports, []int{report_type_received_transaction,
		report_type_transaction_created,
		report_type_block_mined,
		report_type_received_block,
		report_type_blockchain_updated,
		report_type_entire_blockchain,
		report_type_connections})

	max_neighbours := 3
	bootstrap_address := Address{Port: 8080}

	peer_config := PeerConfig{
		Self_Address:          Address{Port: 8080},
		Trailing_Zeros:        20,
		Is_Bootstrap:          true,
		Is_Miner:              false,
		Is_Transaction_Maker:  false,
		Bootstrap_Address:     Address{},
		Transaction_Per_Block: 4,
		Max_Neighbours:        max_neighbours,
		Die_After:             int64(-1),
		Up_Channel:            reports,
	}

	go peer_main(peer_config)

	peer_config.Bootstrap_Address = bootstrap_address
	peer_config.Is_Bootstrap = false

	for i := 8081; i <= 8086; i++ {
		peer_config.Is_Miner = i%2 == 1
		peer_config.Is_Transaction_Maker = i%2 == 0
		peer_config.Self_Address = Address{Port: uint16(i)}

		go peer_main(peer_config)
	}

	// bad node
	peer_config.Is_Miner = true
	peer_config.Self_Address = Address{Port: 8087}
	peer_config.Is_Bad_Node = true
	go peer_main(peer_config)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
