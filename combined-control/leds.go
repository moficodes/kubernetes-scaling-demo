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

func processPixelsForLed(p PixelGrid, mapping []int) []byte {
	pixels := make([]int, len(mapping))

	count := 0
	for i := 0; i < len(p.Pixels); i++ {
		for j := 0; j < len(p.Pixels[i]); j++ {
			pixels[mapping[count]] = p.Pixels[i][j]
			count++
		}
	}

	var res []byte
	for i := 0; i < len(pixels); i++ {
		res = append(res, intToThreeBytes(pixels[i])...)
	}
	return res
}

func processInstancesForLed(mapping []int, board [][]int, cloudRunInstances, gkeInstances map[string]Instance) []byte {

	cloudRunKeys := make([]string, len(cloudRunInstances))
	i := 0
	for k := range cloudRunInstances {
		cloudRunKeys[i] = k
		i++
	}

	gkeKeys := make([]string, len(gkeInstances))
	i = 0
	for k := range gkeInstances {
		gkeKeys[i] = k
		i++
	}

	sort.Strings(cloudRunKeys)
	sort.Strings(gkeKeys)

	counter := 0

	for i := len(board)/2 - 1; i >= 0; i-- {
		for j := 0; j < len(board[i]); j++ {
			if counter < len(cloudRunKeys) {
				instance := cloudRunInstances[cloudRunKeys[counter]]
				if instance.Status == ACTIVE {
					board[i][j] = 0x003800
				} else if instance.Status == IDLE {
					board[i][j] = 0x686000
				} else if instance.Status == TERMINATED {
					board[i][j] = 0x680000
				}
				counter++
			} else {
				board[i][j] = 0x000000
			}
		}
	}

	for i := len(board) - 1; i >= len(board)/2; i-- {
		for j := 0; j < len(board[i]); j++ {
			if counter < len(gkeKeys) {
				instance := gkeInstances[gkeKeys[counter]]
				if instance.Status == ACTIVE {
					board[i][j] = 0x003800
				} else if instance.Status == IDLE {
					board[i][j] = 0x686000
				} else if instance.Status == TERMINATED {
					board[i][j] = 0x680000
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
