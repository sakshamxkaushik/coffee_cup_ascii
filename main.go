package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Particle struct {
	Lifetime int64
	Speed    float64

	X float64
	Y float64
}

type ParticleParams struct {
	MaxLife       int64
	MaxSpeed      float64
	ParticleCount int
	X             int
	Y             int
	Scale         float64

	nextPosition NextPosition
	ascii        Ascii
	reset        Reset
}

type NextPosition func(particle *Particle, deltaMs int64)
type Ascii func(row, col int, count [][]int) string
type Reset func(particle *Particle, params *ParticleParams)

type ParticleSystem struct {
	ParticleParams
	particles []*Particle

	lastTime int64
}

func NewParticleSystem(params ParticleParams) ParticleSystem {
	particles := make([]*Particle, 0)

	for i := 0; i < params.ParticleCount; i++ {
		particles = append(particles, &Particle{})
	}

	return ParticleSystem{
		ParticleParams: params,
		lastTime:       time.Now().UnixMilli(),
		particles:      particles,
	}
}

func (ps *ParticleSystem) Start() {
	for _, p := range ps.particles {
		ps.reset(p, &ps.ParticleParams)
	}
}

func (ps *ParticleSystem) Update() {
	now := time.Now().UnixMilli()
	delta := now - ps.lastTime
	ps.lastTime = now

	for _, p := range ps.particles {
		ps.nextPosition(p, delta)

		if p.Y >= float64(ps.Y) || p.X >= float64(ps.X) || p.Lifetime <= 0 {
			ps.reset(p, &ps.ParticleParams)
		}
	}
}

func (ps *ParticleSystem) Display() string {
	counts := make([][]int, 0)

	for row := 0; row < ps.Y; row++ {
		count := make([]int, 0)
		for col := 0; col < ps.X; col++ {
			count = append(count, 0)
		}
		counts = append(counts, count)
	}

	for _, p := range ps.particles {
		row := int(math.Floor(p.Y))
		col := int(math.Round(p.X))

		counts[row][col]++
	}

	out := make([][]string, 0)
	for r, row := range counts {
		outRow := make([]string, 0)
		for c := range row {
			outRow = append(outRow, ps.ascii(r, c, counts))
		}

		out = append(out, outRow)
	}

	reverse(out)
	outStr := make([]string, 0)

	for _, row := range out {
		outStr = append(outStr, strings.Join(row, ""))
	}

	return strings.Join(outStr, "\n")
}

func reverse(arr [][]string) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

type Coffee struct {
	ParticleSystem
}

var dirs = [][]int{
	{-1, -1},
	{-1, 0},
	{-1, 1},
	{0, -1},
	{0, 1},
	{1, 0},
	{1, 1},
	{1, -1},
}

func countParticles(row, col int, counts [][]int) int {
	count := 0

	for _, dir := range dirs {
		r := row + dir[0]
		c := col + dir[1]
		if r < 0 || r >= len(counts) || c < 0 || c >= len(counts[0]) {
			continue
		}
		count = counts[row+dir[0]][col+dir[1]]
	}

	return count
}

func normalize(row, col int, counts [][]int) {
	if countParticles(row, col, counts) > 4 {
		counts[row][col] = 0
	}
}

func reset(p *Particle, params *ParticleParams) {
	p.Lifetime = int64(math.Floor(float64(params.MaxLife) * rand.Float64()))
	p.Speed = params.MaxSpeed * rand.Float64()

	maxX := math.Floor(float64(params.X) / 2)
	x := math.Max(-maxX, math.Min(rand.NormFloat64()*params.Scale, maxX))

	p.X = x + maxX
	p.Y = 0
}

func nextPos(p *Particle, deltaMs int64) {
	p.Lifetime -= deltaMs

	if p.Lifetime <= 0 {
		return
	}

	percent := (float64(deltaMs) / 2000.0)
	p.Y += p.Speed * percent
}

func NewCoffee(width, height int, scale float64) Coffee {
	if width%2 == 0 {
		log.Fatal("width must be odd number")
	}

	ascii := func(row, col int, counts [][]int) string {
		count := counts[row][col]

		if count == 0 {
			return " "
		}
		if count < 4 {
			return "░"
		}
		if count < 6 {
			return "▒"
		}
		if count < 9 {
			return "▓"
		}

		return "█"
	}

	return Coffee{
		ParticleSystem: NewParticleSystem(
			ParticleParams{
				MaxLife:       7000,
				MaxSpeed:      1.5,
				ParticleCount: 60,

				reset:        reset,
				ascii:        ascii,
				nextPosition: nextPos,
				X:            width,
				Y:            height,
				Scale:        scale,
			},
		),
	}
}

var cup = `
                    .:-----====----------------:.                     
                 .:=-===++--:::-===========+=------:                  
                ::==+===-==:::::-:.:--:--:::--===----:                
               .:++===:..--:::::::::::..--....:-==--+.:               
               .=:=+==-::.::::::::::::::-:.....:===-:-=::....:        
                :.:--=+==--::::...........:::-==+=-----. .... :.      
              :-=   :::---==+++==------==++==---:::..=..:   .: -      
          .-=-:.-        .....::--------::.....   ...=.+-:  :: -      
        .=-:.....-                                ..=.=..:+-. :.      
       -=:.......:-                              ..-::----. :=        
      -=........  .-                            ..-:  .  ::-.=-       
      -=......      :.                          :-.::::::.....-:      
      --:....        .:.                      :-.::.     ....:-:      
       ---..           ==:                  :=-          ...:--       
        .---.           .-=-::.        .::-=-.          ..:--:        
          -==-:.           .::--======--::.            ::-=-          
            :-=-=-::.                             .::-=-=-            
               ::-=:----::.......     .......::-----=-:.              
                   .::::--:::::---------:::::--::::.                  
                          ...::::::::::::....                        `

func main() {
	coffee := NewCoffee(71, 8, 4.5)
	coffee.Start()

	timer := time.NewTicker(100 * time.Millisecond)
	for {
		<-timer.C
		fmt.Print("\033[H\033[2J")
		coffee.Update()
		fmt.Print(coffee.Display())
		fmt.Print(cup)
	}
}
