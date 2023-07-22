package main

import (
	"sort"
)

const (
	RED    = 0xFF0000
	GREEN  = 0x00FF00
	YELLOW = 0xFFEA00
)

func singleBoardMapping(rows, cols, position int) [][]int {
	var board [][]int
	for i := 0; i < rows; i++ {
		board = append(board, make([]int, cols))
	}

	value := rows*cols - 1
	max := rows * cols

	for i := 0; i < rows; i++ {
		if i%2 == 0 {
			for j := 0; j < cols; j++ {
				board[i][j] = value + (position * max)
				value--
			}
		} else {
			for j := cols - 1; j >= 0; j-- {
				board[i][j] = value + (position * max)
				value--
			}
		}
	}
	return board
}

func boardMapping(rows, cols int, positions [][]int) []int {
	var board [][]int
	var result []int

	width := len(positions)
	height := len(positions[0])
	for i := 0; i < rows*height; i++ {
		board = append(board, make([]int, cols*width))
	}

	for i := 0; i < len(positions); i++ {
		for j := 0; j < len(positions[i]); j++ {
			sbm := singleBoardMapping(rows, cols, positions[i][j])
			for x := 0; x < len(sbm); x++ {
				for y := 0; y < len(sbm[x]); y++ {
					board[i*rows+x][j*cols+y] = sbm[x][y]
				}
			}
		}
	}

	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			result = append(result, board[i][j])
		}
	}
	return result
}

func intToThreeBytes(i int) []byte {
	return []byte{byte(i >> 16), byte(i >> 8), byte(i)}
}

func processInstancesForLed(mapping []int, board [][]int, instances map[string]Instance) []byte {

	keys := make([]string, len(instances))
	i := 0
	for k := range instances {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	counter := 0

	for i := len(board) - 1; i >= 0; i-- {
		for j := 0; j < len(board[i]); j++ {
			if counter < len(keys) {
				instance := instances[keys[counter]]
				if instance.Status == ACTIVE {
					board[i][j] = 0x00FF00
				} else if instance.Status == IDLE {
					board[i][j] = 0xFFEA00
				} else if instance.Status == TERMINATED {
					board[i][j] = 0xFF0000
				}
				counter++
			} else {
				board[i][j] = 0x000000
			}
		}
	}

	resInt := make([]int, len(mapping))

	count := 0
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[i]); j++ {
			resInt[mapping[count]] = board[i][j]
			count++
		}
	}

	var res []byte
	for i := 0; i < len(resInt); i++ {
		res = append(res, intToThreeBytes(resInt[i])...)
	}

	return res
}
