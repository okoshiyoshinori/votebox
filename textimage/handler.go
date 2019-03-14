package textimage

import (
	"log"

	"github.com/fogleman/gg"

	"github.com/okoshiyoshinori/votebox/config"
)

func MakeOGP(t string, b *Base) {
	im, err := gg.LoadImage(config.APConfig.Ogb + "base.jpeg")
	if err != nil {
		log.Println(err)
	}
	dc := gg.NewContext(b.Width, b.Height)
	dc.DrawImage(im, 0, 0)
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("font/azuki_new.ttf", 36); err != nil {
		log.Println(err)
	}
	dc.DrawStringWrapped(wordWrapUtil(t, 33), 600, 300, 0.5, 0.5, 900, 2, gg.AlignCenter)
	dc.SavePNG(config.APConfig.Ogb + b.OutPath + ".jpeg")
}
