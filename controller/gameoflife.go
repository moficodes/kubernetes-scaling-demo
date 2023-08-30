package main

func nextGeneration(board [][]int) [][]int {
	if len(board) == 0 || len(board[0]) == 0 {
		return board
	}

	rows, cols := len(board), len(board[0])
	newBoard := make([][]int, rows)
	for i := range newBoard {
		newBoard[i] = make([]int, cols)
	}

	// Check the number of live neighbors for each cell
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			liveNeighbors, color := calculateLiveNeighborsColor(board, i, j)

			// Rule 1 or Rule 3
			if board[i][j] != 0 && (liveNeighbors < 2 || liveNeighbors > 3) {
				newBoard[i][j] = 0
			} else if board[i][j] != 0 && (liveNeighbors == 2 || liveNeighbors == 3) {
				// Rule 2
				newBoard[i][j] = color
			} else if board[i][j] == 0 && liveNeighbors == 3 {
				// Rule 4
				newBoard[i][j] = color
			}
		}
	}

	return newBoard
}

// Helper function to count live neighbors for a given cell
func calculateLiveNeighborsColor(board [][]int, x int, y int) (int, int) {
	directions := [][]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	count := 0
	r, g, b := 0, 0, 0
	for _, dir := range directions {
		newX, newY := x+dir[0], y+dir[1]
		if newX >= 0 && newX < len(board) && newY >= 0 && newY < len(board[0]) && board[newX][newY] != 0 {
			count++
			r += board[newX][newY] >> 16 & 0xFF
			g += board[newX][newY] >> 8 & 0xFF
			b += board[newX][newY] & 0xFF
		}
	}
	if count == 0 {
		return count, 0
	}

	color := (r/count)<<16 + (g/count)<<8 + b/count

	return count, color
}
