package textimage

var (
	Sub = &Subscribe{
		Question: make(chan map[string]string),
	}
	Worker = new(Maker)
	base   = &Base{
		Width:    1200,
		Height:   630,
		Quality:  100,
		FontPath: "font/azuki.ttf",
	}
)

type Maker struct {
	H Handler
}

func (m *Maker) setHandle(hd Handler) {
	m.H = hd
}

func (m *Maker) make(t string, b *Base) {
	m.H.MakeImage(t, b)
}

type Base struct {
	Width    int
	Height   int
	Quality  int
	FontPath string
	OutPath  string
}

type Subscribe struct {
	Question chan map[string]string
}

type Handler interface {
	MakeImage(string, *Base)
}

type HandleFunc func(t string, B *Base)

func (h HandleFunc) MakeImage(t string, b *Base) {
	h(t, b)
}

func NewMaker(h Handler) *Maker {
	return &Maker{}
}

func init() {
	Worker.setHandle(HandleFunc(MakeOGP))
}

func (s *Subscribe) Run() {
	for {
		select {
		case message := <-s.Question:
			for k, v := range message {
				base.OutPath = k
				Worker.make(v, base)
			}
		}
	}
}
