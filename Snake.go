package main

import (
    "time"
    "math"
    "math/rand"
    "os"
    "image"
    "image/color"
    "fmt"
    //"path"
    //"strings"

    _ "image/png"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/imdraw"
    "github.com/faiface/pixel/pixelgl"
    "github.com/faiface/pixel/text"
    "golang.org/x/image/colornames"
    "golang.org/x/image/font/basicfont"
)

const (
    TILE_SIZE = 60
)

var (
    win *pixelgl.Window
    imd *imdraw.IMDraw
    TILES pixel.Vec
    Center pixel.Vec
    Food []pixel.Vec
)

const (
    UP = iota
    DOWN
    RIGHT
    LEFT
    NODIR
)

type Snake struct {
    Pieces []pixel.Vec
    Direction uint8
    MoveQueue []uint8
    Adding bool
}

func S() Snake {
    return Snake{Pieces: []pixel.Vec{Center}, Direction: NODIR, MoveQueue: []uint8{}, Adding: false}
}

func (s *Snake) Step() bool {
    Head := s.Pieces[len(s.Pieces)-1]

    if len(s.MoveQueue) > 0 {
        s.Direction = s.MoveQueue[0]
        s.MoveQueue = s.MoveQueue[1:]
    }

    switch s.Direction {
    case UP:
        Position := pixel.V(Head.X, Head.Y+1)
        if Position.Y >= TILES.Y {
            return true
        }
        s.Pieces = append(s.Pieces, Position)
    case DOWN:
        Position := pixel.V(Head.X, Head.Y-1)
        if Position.Y < 0 {
            return true
        }
        s.Pieces = append(s.Pieces, Position)
    case RIGHT:
        Position := pixel.V(Head.X+1, Head.Y)
        if Position.X >= TILES.X {
            return true
        }
        s.Pieces = append(s.Pieces, Position)
    case LEFT:
        Position := pixel.V(Head.X-1, Head.Y)
        if Position.X < 0 {
            return true
        }
        s.Pieces = append(s.Pieces, Position)
    }
    if !s.Adding && s.Direction != NODIR {
        s.Pieces = s.Pieces[1:]
    }
    s.Adding = false
    return s.HitSelf()
}

func Opposite(Current uint8, Next uint8) bool {
    return (Current == UP && Next == DOWN) || (Current == DOWN && Next == UP) || (Current == RIGHT && Next == LEFT) || (Current == LEFT && Next == RIGHT)
}

func (s *Snake) Draw() {
    imd.Color = colornames.White
    for _, Piece := range s.Pieces {
        BottomLeft := pixel.V((Piece.X*TILE_SIZE)+2, (Piece.Y*TILE_SIZE)+2)
        imd.Push(BottomLeft)
        UpperRight := pixel.V((Piece.X*TILE_SIZE)+(TILE_SIZE-4), (Piece.Y*TILE_SIZE)+(TILE_SIZE-4))
        imd.Push(UpperRight)
        imd.Rectangle(0)
    }
    if len(s.Pieces) > 0 {
        s.DrawFace()
    }
}

func (s *Snake) DrawFace() {
    imd.Color = colornames.Gray
    var toleft, fromleft, toright, fromright pixel.Vec
    switch s.Direction {
    case LEFT:
        fromleft = pixel.V(4, 4)
        toleft = pixel.V(10, 6).Add(fromleft)
        
        fromright = pixel.V(4, TILE_SIZE-9)
        toright = fromright.Sub(pixel.V(-10, 6))
    case RIGHT:
        fromleft = pixel.V(TILE_SIZE-19, 4)
        toleft = pixel.V(10, 6).Add(fromleft)
        
        fromright = pixel.V(TILE_SIZE-19, TILE_SIZE-9)
        toright = fromright.Sub(pixel.V(-10, 6))
    case UP:
        fromleft = pixel.V(4, TILE_SIZE-19)
        toleft = pixel.V(6, 10).Add(fromleft)
        
        fromright = pixel.V(TILE_SIZE-9, TILE_SIZE-19)
        toright = fromright.Sub(pixel.V(6, -10))
    case DOWN:
        fromleft = pixel.V(4, 4)
        toleft = pixel.V(6, 10).Add(fromleft)
        
        fromright = pixel.V(TILE_SIZE-9, 4)
        toright = fromright.Sub(pixel.V(6, -10))
    }

   start := pixel.V((s.Pieces[len(s.Pieces)-1].X*TILE_SIZE)+2, (s.Pieces[len(s.Pieces)-1].Y*TILE_SIZE)+2)

   imd.Push(start.Add(fromleft), start.Add(toleft))
   imd.Rectangle(0)

   imd.Push(start.Add(fromright), start.Add(toright))
   imd.Rectangle(0)
}

func (s *Snake) Add() {
    s.Adding = true
}

func (s *Snake) Intersects(Position pixel.Vec) bool {
    for _, Pos := range s.Pieces {
        if Pos.Eq(Position) {
            return true
        }
    }
    return false
}

func (s *Snake) HitSelf() bool {
    Head := s.Pieces[len(s.Pieces)-1]
    for i := len(s.Pieces)-2; i >= 0; i-- {
        Pos := s.Pieces[i]
        if Pos.Eq(Head) {
            return true
        }
    }
    return false
}

func DrawFood() {
    imd.Color = colornames.Lime
    for _, pos := range Food {
        BottomLeft := pos.Scaled(TILE_SIZE).Add(pixel.V(2, 2))
        UpperRight := pos.Scaled(TILE_SIZE).Add(pixel.V(TILE_SIZE-4, TILE_SIZE-4))
        imd.Push(BottomLeft)
        imd.Push(UpperRight)
        imd.Rectangle(0)
    }
}

func Reset(s *Snake) {
    time.Sleep(time.Second/2)
    for i := len(s.Pieces)-1; i >= 0; i-- {
        win.Clear(colornames.Black)
        imd.Clear()
        s.Pieces = s.Pieces[0:i]
        DrawFood()
        s.Draw()
        imd.Draw(win)
        win.Update()
        time.Sleep(time.Second/20)
    }
    *s = S()
    Food = []pixel.Vec{}
}

func LoadPicture(path string) (pixel.Picture, error) {
    file, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    img, _, err := image.Decode(file)
    if err != nil {
        panic(err)
    }
    return pixel.PictureDataFromImage(img), nil
}

func run() {
    rand.Seed(time.Now().Unix())

    monitor := pixelgl.PrimaryMonitor()
    PositionX, PositionY  := monitor.Position()
    SizeX, SizeY := monitor.Size()
    screen := pixel.R(PositionX, PositionY, SizeX, SizeY)

    //Only works with an exe
    //ImgPath := path.Join(path.Dir(strings.ReplaceAll(os.Args[0], "\\", "/")), "Snake.png") //Only works with an exe ("go build")
    ImgPath := "Snake.png" //Only works with "go run"

    icon, err := LoadPicture(ImgPath)
    if err != nil {
        panic(err)
    }

    cfg := pixelgl.WindowConfig{
        Title:   "Snake",
        Monitor: pixelgl.PrimaryMonitor(),
        Bounds:  screen,
        Icon: []pixel.Picture{icon},
    }
    win, err = pixelgl.NewWindow(cfg)
    if err != nil {
        panic(err)
    }

    TILES = pixel.V(math.Floor(win.Bounds().W()/TILE_SIZE), math.Floor(win.Bounds().H()/TILE_SIZE))
    Center = TILES.Scaled(0.5).Floor()
    imd = imdraw.New(nil)
    
    fps := time.NewTicker(time.Second/60)
    update := time.NewTicker(time.Second/6)
    defer fps.Stop()
    defer update.Stop()

    atlas := text.NewAtlas(
        basicfont.Face7x13,
        []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'},
    )
    txt := text.New(pixel.V(20, 20), atlas)
    txt.Color = color.RGBA{105, 105, 105, 175}
    Dot := win.Bounds().Max.Sub(pixel.V(txt.BoundsOf("0").W()*5, txt.BoundsOf("0").H()*10))

    snake := S()

    for !win.Closed() {
        win.Clear(colornames.Black)
        imd.Clear()

        if win.JustPressed(pixelgl.KeyEscape) {
            win.SetClosed(true)
        }

        if win.JustPressed(pixelgl.KeyUp) && (len(snake.MoveQueue) == 0 || snake.MoveQueue[len(snake.MoveQueue)-1] != UP) {
            var Next uint8
            if len(snake.MoveQueue) == 0 {
                Next = snake.Direction
            } else {
                Next = snake.MoveQueue[len(snake.MoveQueue)-1]
            }
            if !Opposite(Next, UP) {
                snake.MoveQueue = append(snake.MoveQueue, UP)
            }
        }
        if win.JustPressed(pixelgl.KeyDown) && (len(snake.MoveQueue) == 0 || snake.MoveQueue[len(snake.MoveQueue)-1] != DOWN) {
            var Next uint8
            if len(snake.MoveQueue) == 0 {
                Next = snake.Direction
            } else {
                Next = snake.MoveQueue[len(snake.MoveQueue)-1]
            }
            if !Opposite(Next, DOWN) {
                snake.MoveQueue = append(snake.MoveQueue, DOWN)
            }
        }
        if win.JustPressed(pixelgl.KeyRight) && (len(snake.MoveQueue) == 0 || snake.MoveQueue[len(snake.MoveQueue)-1] != RIGHT) {
            var Next uint8
            if len(snake.MoveQueue) == 0 {
                Next = snake.Direction
            } else {
                Next = snake.MoveQueue[len(snake.MoveQueue)-1]
            }
            if !Opposite(Next, RIGHT) {
                snake.MoveQueue = append(snake.MoveQueue, RIGHT)
            }
        }
        if win.JustPressed(pixelgl.KeyLeft) && (len(snake.MoveQueue) == 0 || snake.MoveQueue[len(snake.MoveQueue)-1] != LEFT) {
            var Next uint8
            if len(snake.MoveQueue) == 0 {
                Next = snake.Direction
            } else {
                Next = snake.MoveQueue[len(snake.MoveQueue)-1]
            }
            if !Opposite(Next, LEFT) {
                snake.MoveQueue = append(snake.MoveQueue, LEFT)
            }
        }

        select {
        case <-update.C:
            if len(Food) < 3 {
                NewPiece := pixel.V(float64(rand.Intn(int(TILES.X))), float64(rand.Intn(int(TILES.Y))))
                for snake.Intersects(NewPiece) {
                    NewPiece = pixel.V(float64(rand.Intn(int(TILES.X))), float64(rand.Intn(int(TILES.Y))))
                }
                Food = append(Food, NewPiece)
            }
            HitSelf := snake.Step()
            if HitSelf {
                Reset(&snake)
            }
            for Index, Piece := range Food {
                if snake.Intersects(Piece) {
                    snake.Add()
                    Food = append(Food[0:Index], Food[Index+1:]...)
                }
            }
        default:
            //Make this check non-blocking
        }

        Score := len(snake.Pieces)-1
        if snake.Adding {
            Score++
        }

        txt.Clear()
        txt.Dot = Dot
        txt.WriteString(fmt.Sprint(Score))
        DrawFood()
        snake.Draw()

        imd.Draw(win)
        txt.Draw(win, pixel.IM.Scaled(txt.Dot, 10))
        win.Update()
        <- fps.C
    }
}

func main() {
    pixelgl.Run(run)
}