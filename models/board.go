package models

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Model of the Board
type Board struct {
  Blocks [25][15]Block
  IsFalling bool
  FigureOptions []Figure
  CurrentFigure Figure
  FigureFallingId int
}

// This method only prints the Blocks with Active == True
func (board *Board) PrintBoard () {
  for _, bl := range board.Blocks {
    fmt.Print("|")
    for _, b := range bl {
      if (b.Active) {
        fmt.Print(" 󰝣 ")
      }else{ fmt.Print("   ") }
    }
    fmt.Print("|\n")
  }
}

// This method sets the attribute active of the falling block to false in the board before the new iteration
// and return a backup of the current position of the figure if something goes wrong.
func (board *Board) clean () [][]int{
  cBak :=[][]int{}
  for _, coordinate := range board.CurrentFigure.Coordinates {
    cBak = append(cBak, coordinate)
    board.Blocks[coordinate[0]][coordinate[1]].Active = false
    board.Blocks[coordinate[0]][coordinate[1]].FigureId = board.FigureFallingId - 1

  }
  return cBak
}

// the method responsible of changing the position of the falling figure
func (board *Board) RunFrame(mutex *sync.Mutex) {
  mutex.Lock()
  if ( !board.IsFalling ) { board.InitFigures() }
  var coordinates [][]int
  var coordinates_bak [][]int
  success := true
  coordinates_bak = board.clean()
  for _, coordinate := range board.CurrentFigure.Coordinates {
    if coordinate[0] + 1  > len(board.Blocks) - 1 {
      success = false
      break
    }
    if board.Blocks[coordinate[0] + 1][coordinate[1]].FigureId != board.FigureFallingId && !board.Blocks[coordinate[0] + 1][coordinate[1]].Active{
      val := coordinate[0]
      val2 := coordinate[1]
      coordinates = append(coordinates, []int{val + 1, val2})
    }else{
      success = false
    }
  }

  if success{
    board.CurrentFigure.Coordinates = coordinates
    for _, coordinate := range coordinates {
      board.Blocks[coordinate[0]][coordinate[1]].Active = true
      board.Blocks[coordinate[0]][coordinate[1]].FigureId = board.FigureFallingId
    }
  }else {
    board.CurrentFigure.Coordinates = coordinates_bak
    for _, coordinate := range coordinates_bak {
      board.Blocks[coordinate[0]][coordinate[1]].Active = true
      board.Blocks[coordinate[0]][coordinate[1]].FigureId = board.FigureFallingId - 1
    }
    board.IsFalling = false
    board.CurrentFigure = Figure{}
  }
  mutex.Unlock()
}

// Init a new figure at the top of the board
func (b *Board) InitFigures()  {
  b.FigureFallingId ++
  b.FigureOptions = []Figure{
    { [][]int{ {0, 3}, {0, 4}, {0, 5}, {1, 4} } },
    { [][]int{ {0, 3}, {0, 4}, {1, 3}, {1, 4} } },
    { [][]int{ {0, 3}, {1, 3}, {2, 3} } },
  }
  figureIndex := rand.Intn((len(b.FigureOptions)))
  b.CurrentFigure = b.FigureOptions[figureIndex]
  b.IsFalling = true
}

func (board *Board) Init()  {
  board.IsFalling = false
  board.FigureFallingId = 0
}

// change the position of the figure to left or right depending of the pressed key (use a gorutine to avoid blocking the game)
func (board *Board) Movement(wg *sync.WaitGroup, mutex *sync.Mutex)  {
  for {
    var key string = ""
    fmt.Scanln(&key)
    movement := 0
    success := true
    if key == "d" { movement = 1 }
    if key == "a" { movement = -1 }
    mutex.Lock()
    for _, coordinate := range board.CurrentFigure.Coordinates {
      if coordinate[1] + movement > len(board.Blocks[0]) - 1 || coordinate[1] + movement < 0 {
        success = false
      }
    }
    if success {
      for _, coordinate := range board.CurrentFigure.Coordinates {
        board.Blocks[coordinate[0]][coordinate[1]].Active = false
        board.Blocks[coordinate[0]][coordinate[1]].FigureId -= 1
        coordinate[1] += movement
      }
    }
    mutex.Unlock()
  }
}


func (board *Board) eraseLine(mutex *sync.Mutex) {
  mutex.Lock()
  for i, line := range board.Blocks {
    isBlock := true
    for _, block := range line{
      if !block.Active || block.FigureId == board.FigureFallingId {
        isBlock = false
        break;
      }
    }
    if isBlock {
      board.IsFalling = false
      for j := range line{
        board.Blocks[i][j].Active = false
        board.Blocks[i][j].FigureId -= 1
      }
      board.InitFigures()
    }
  }
  mutex.Unlock()
}

func (board *Board) Run () {
  var wg sync.WaitGroup
  var mutex sync.Mutex
  board.Init()
  go board.Movement(&wg, &mutex)
  for {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
    board.eraseLine(&mutex)
    board.RunFrame(&mutex)
    board.PrintBoard()
    time.Sleep(time.Millisecond * 180)
  }
}
