package entities

type Player struct {
	Name      string
	Input     chan string
	output    chan string
	Location  string
	Wearing   []string
	Inventory []string
}

func NewPlayer(name string) *Player {
	return &Player{
		Name:      name,
		Input:     make(chan string, 10),
		output:    make(chan string, 10),
		Location:  "кухня",
		Wearing:   []string{},
		Inventory: []string{},
	}
}

func (p *Player) GetOutput() <-chan string {
	return p.output
}

func (p *Player) HandleInput(command string) {
	p.Input <- command
}

func (p *Player) SendMessage(message string) {
	select {
	case p.output <- message:
	default:
	}
}
