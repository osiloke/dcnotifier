package main

import (
	"time"
	"log"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/firmata"
	"github.com/hybridgroup/gobot/platforms/gpio" 
)
 
type note struct {
	tone     float64
	duration float64
}

type Song struct{
	name string
	notes []note
	repeat int
	done chan bool
	stopped bool
}
func (s *Song) stop(){ 
	s.done <- true
}

func NewSong(name string, notes []note, repeat int) *Song{
	return &Song{name, notes,repeat, make(chan bool), true}
}
func songPlayer(buzzer *gpio.BuzzerDriver, songs chan *Song){
	var lastSong *Song
	for {
	    select{
		case song := <- songs:
			log.Println("received song "+song.name)
			if lastSong != nil{ 
				if !lastSong.stopped{
					lastSong.stop()
				} 
			} 
			lastSong = song
			go func(){
				i := 0
				cnt := len(song.notes) - 1
				repeat := song.repeat
				song.stopped = false
outer:
				for{  
					select{
					case <- song.done:
						log.Println("song done triggered "+song.name)
						break outer
					
					case <-time.After(50*time.Millisecond):
						note := song.notes[i]
						buzzer.Tone(note.tone, note.duration)
						log.Printf("playing note %d from %s", i+1, song.name) 
						i++
						if i >= cnt{
							log.Println("notes finished "+song.name)
							repeat--
							if repeat <= 0{  
								break outer								
							}
							time.Sleep(1000*time.Millisecond)
							i = 0
						}  
					}	
				}
				  
				song.stopped = true
				log.Println("stopped player for "+song.name)
			}()
		}
	}
	log.Println("ending song player")
}
func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("arduino", "/dev/ttyACM0")
	buzzer := gpio.NewBuzzerDriver(firmataAdaptor, "buzzer", "3")
	button := gpio.NewButtonDriver(firmataAdaptor, "poweroff", "2")	

	work := func() {

		song := []note{
			{gpio.Gb0, gpio.Eighth},
			{gpio.C1, gpio.Eighth},
			{gpio.Rest, gpio.Whole},
		}  
		song2 := []note{ 
			{gpio.C0, gpio.Half},
//			{gpio.Ab1, gpio.Eighth},
			{gpio.Rest, gpio.Whole},
		}  
		songs := make(chan *Song)
		go songPlayer(buzzer, songs)
		gobot.On(button.Event("release"), func(data interface{}) {
			log.Println("power on") 
			songs <- NewSong("power on", song2, 1)
	    })
		gobot.On(button.Event("push"), func(data interface{}) {
			log.Println("power off")  
			songs <- NewSong("power off", song, 2)
	    })		
	}

	robot := gobot.NewRobot("powerDetector",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{buzzer, button},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
