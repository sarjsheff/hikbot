package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sarjsheff/hiklib"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

var ipParam = flag.String("c", "", "Camera IP address.")
var userParam = flag.String("u", "", "Username.")
var passParam = flag.String("p", "", "Password.")
var tkeyParam = flag.String("t", "", "Telegram key.")
var adminParam = flag.Int("a", 0, "Telegram userid.")
var alarmParam = flag.Int("b", 7200, "Alarm port.")

var datadirParam = flag.String("d", "/tmp", "Data dir, default: /tmp .")
var previewsizeParam = flag.Int("s", 20000000, "Video preview byte size.")
var zParam = flag.Int("z", 2, "Video preview rescale (divide).")

var x1Param = flag.Bool("x1", false, "Issue 1.")

// type AlarmItem struct {
// 	IP        string
// 	Command   int
// 	AlarmType int
// }

type FFProbe struct {
	Streams []map[string]interface{} `json:"streams"`
	Format  map[string]interface{}   `json:"format"`
}

var appid uuid.UUID = uuid.Must(uuid.NewRandom())

var motions chan hiklib.AlarmItem

// var dev = C.DevInfo{byStartChan: 0}
// var user = C.int(-1)
var dev = hiklib.DevInfo{ByStartChan: 0}
var user = -1

type touser string
type MotionArea struct{ x, y, w, h float32 }

func (t touser) Recipient() string {
	return string(t)
}

// //export onmessagev30
// func onmessagev30(command C.int, pAlarmer *C.NET_DVR_ALARMER, pAlarmInfo *C.char, dwBufLen C.uint, pUserData unsafe.Pointer) {
// 	i := AlarmItem{IP: C.GoString(&pAlarmer.sDeviceIP[0]), Command: int(command)}
// 	switch int(command) {
// 	case COMM_ALARM_V30:
// 		log.Println("ALARM")
// 		i.AlarmType = int(C.getalarminfo(pAlarmInfo).dwAlarmType)
// 		motions <- i
// 		break
// 	case COMM_DEV_STATUS_CHANGED:
// 		log.Printf("COMM_DEV_STATUS_CHANGED")
// 		break
// 	default:
// 		log.Printf("Unknown Alarm [0x%x] !!!", command)
// 	}
// }

// //export onmessage
// func onmessage(command C.int, ip *C.char, data *C.char, ln C.uint) C.int {

// 	i := AlarmItem{IP: C.GoString(ip), Command: int(command)}

// 	switch int(command) {
// 	case COMM_ALARM_V30:
// 		i.AlarmType = int(C.getalarminfo(data).dwAlarmType)
// 		motions <- i
// 		break
// 	case COMM_DEV_STATUS_CHANGED:
// 		log.Printf("COMM_DEV_STATUS_CHANGED %s %s", C.GoString(ip), C.GoString(data))
// 		break
// 	default:
// 		log.Printf("Unknown Alarm [0x%x] %s %s !!!", command, C.GoString(ip), C.GoString(data))
// 	}
// 	return 1
// }

func bot() {

	videolist := map[string]string{}

	done := make(chan int, 1)
	done <- 1

	admin := touser(strconv.Itoa(*adminParam))
	b, err := tb.NewBot(tb.Settings{
		Token:  *tkeyParam,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	// var menu = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	// var btnSettings = menu.Data("⚙", "Settings")
	video := func(offset int, limit int) {
		//var v = hiklib.MotionVideos{} //C.MotionVideos{}
		mm, _ := b.Send(admin, "Fetch video from camera...")
		//C.HListVideo(C.int(user), &v)
		_, v := hiklib.HikListVideo(user)
		b.Edit(mm, strconv.Itoa(v.Count)+" video on camera.")
		if v.Count > 0 {
			txt := ""
			if offset == 0 {
				txt = fmt.Sprintf("First %d video:\n", limit)
			} else {
				txt = fmt.Sprintf("%d video from %d :\n", limit, offset)
			}
			for i := offset - 1; i < v.Count && i < offset+limit-1; i++ {
				dt := time.Date(v.Videos[i].From_year, time.Month(v.Videos[i].From_month), v.Videos[i].From_day, v.Videos[i].From_hour, v.Videos[i].From_min, v.Videos[i].From_sec, 0, time.UTC)
				todt := time.Date(v.Videos[i].To_year, time.Month(v.Videos[i].To_month), v.Videos[i].To_day, v.Videos[i].To_hour, v.Videos[i].To_min, v.Videos[i].To_sec, 0, time.UTC)
				txt = txt + "<b>" + dt.Format("2006-01-02 15:04:05") + " - " + todt.Format("15:04:05") + "</b> /dl_" + v.Videos[i].Filename + " \n"
				videolist[v.Videos[i].Filename] = dt.Format("2006-01-02/15:04:05")
			}
			if offset+limit < v.Count {
				txt = txt + fmt.Sprintf("<b>Next 10 video /video_%d_%d</b>\n", offset+limit, limit)
			}
			// menu.Inline(menu.Row(btnSettings))
			// b.Send(admin, txt, &tb.SendOptions{ReplyMarkup: menu, ParseMode: tb.ModeHTML})

			_, err = b.Send(admin, txt, &tb.SendOptions{ParseMode: tb.ModeHTML})
			if err != nil {
				log.Println(err)
			}
		}
	}

	snapshot := func(mareas bool) {
		fname := filepath.Join(*datadirParam, fmt.Sprintf("%s.jpeg", uuid.Must(uuid.NewRandom()).String()))
		err := hiklib.HikCaptureImage(user, dev.ByStartChan, fname)
		if err > -1 {
			caption := ""
			if mareas {
				// var ma = C.MotionAreas{}
				// C.HMotionArea(C.int(user), &ma)
				_, ma := hiklib.HikMotionArea(user)
				col := color.RGBA{255, 0, 0, 128}
				var dst *image.RGBA
				var b image.Rectangle
				f, err := os.Open(fname)
				if err == nil {

					defer f.Close()
					img, _, err := image.Decode(f)
					if err == nil {
						b = img.Bounds()
						dst = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
						draw.Draw(dst, b, img, b.Min, draw.Src)
					}
				}
				caption = caption + fmt.Sprintf("Image size %vx%v\n", b.Dx(), b.Dy())
				for i := 0; i < 8; i++ {
					if ma.Areas[i].W > 0 && ma.Areas[i].H > 0 {
						x, y, w, h := int(float32(b.Dx())*float32(ma.Areas[i].X)), int(float32(b.Dy())*float32(ma.Areas[i].Y)), int(float32(b.Dx())*float32(ma.Areas[i].W)), int(float32(b.Dy())*float32(ma.Areas[i].H))
						log.Printf("Area %v x:%v y:%v [%vx%v]\n", i+1, x, y, w, h)
						caption = caption + fmt.Sprintf("Area %v x:%v y:%v [%vx%v]\n", i+1, x, y, w, h)
						if dst != nil {
							Rect(dst, x, y, w, h, col)
						}
					}
				}
				if dst != nil {
					f.Close()
					f, err = os.Create(fname)
					if err == nil {
						defer f.Close()
						opt := jpeg.Options{
							Quality: 100,
						}
						err = jpeg.Encode(f, dst, &opt)
					}
				}
			}
			//p := &tb.Photo{File: tb.FromDisk(fname)}
			//b.SendAlbum(admin, tb.Album{p})
			p := &tb.Document{File: tb.FromDisk(fname), MIME: "image/jpeg", FileName: time.Now().Format(time.RFC3339) + ".jpeg"}
			if caption != "" {
				p.Caption = caption
			}
			b.Send(admin, p)
			os.Remove(fname)
		} else {
			b.Send(admin, fmt.Sprintf("Error get snapshot [%d].", err))
		}
	}

	// On inline button pressed (callback)
	// b.Handle(&btnSettings, func(c *tb.Callback) {
	// 	b.Respond(c, &tb.CallbackResponse{Text: "testttt"})
	// })

	b.Handle("/video", func(m *tb.Message) {
		<-done
		if m.Sender.ID == *adminParam {
			video(1, 10)
		}
		done <- 1
	})

	b.Handle("/mareas", func(m *tb.Message) {
		<-done
		if m.Sender.ID == *adminParam {
			snapshot(true)
		}
		done <- 1
	})

	b.Handle("/snap", func(m *tb.Message) {
		<-done
		if m.Sender.ID == *adminParam {
			snapshot(false)
		}
		done <- 1
	})

	b.Handle("/reboot", func(m *tb.Message) {
		<-done
		if m.Sender.ID == *adminParam {
			res := hiklib.HikReboot(user)
			if res > 0 {
				b.Send(m.Sender, "Rebooting! Wait 10 sec.")
				time.Sleep(10 * time.Second)
				for Login() < 1 {
					b.Send(m.Sender, "Wait 3 sec.")
					time.Sleep(3 * time.Second)
				}
				b.Send(m.Sender, "Camera online.")
			} else {
				b.Send(m.Sender, fmt.Sprintf("Fail [%d].", res))
			}
		}
		done <- 1
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		<-done
		if m.Sender.ID == *adminParam {
			if strings.HasPrefix(m.Text, "/dl_") {
				mm, _ := b.Send(admin, "Loading...")
				log.Println(m.Text[4:])

				if filename, ok := videolist[m.Text[4:]]; ok {
					os.MkdirAll(filepath.Join(*datadirParam, strings.Split(filename, "/")[0]), 0755)
					fname := filepath.Join(*datadirParam, filename+".mpeg")
					p := &tb.Video{}

					if _, err := os.Stat(fname); os.IsNotExist(err) {
						opts := ffmpeg.KwArgs{
							"format": "mp4",
							//"fs":       strconv.Itoa(*previewsizeParam),
							"vcodec":   "copy", //"libx264",
							"preset":   "ultrafast",
							"acodec":   "none",
							"movflags": "+faststart",
						}

						// C.HSaveFile(C.int(user), C.CString(m.Text[4:]), C.CString(fname))
						hiklib.HikSaveFile(user, m.Text[4:], fname)
						b.Edit(mm, "Probing...")
						f, err := ffmpeg.Probe(fname)
						var fjson FFProbe
						err = json.Unmarshal([]byte(f), &fjson)
						if err == nil {
							// b.Send(admin, f)

							p.Width = int(fjson.Streams[0]["width"].(float64))
							p.Height = int(fjson.Streams[0]["height"].(float64))

							if sz, err := strconv.Atoi(fjson.Format["size"].(string)); err == nil {
								if sz > *previewsizeParam {
									if s, err := strconv.ParseFloat(fjson.Format["duration"].(string), 64); err == nil {

										opts["vcodec"] = "libx264"
										opts["b"] = strconv.Itoa(int(math.Floor(float64(*previewsizeParam)/math.Floor(s)) * 8))
										p.Width = int(math.Round(float64(p.Width) / float64(*zParam)))
										p.Height = int(math.Round(float64(p.Height) / float64(*zParam)))
										opts["vf"] = fmt.Sprintf("scale=%d:%d", p.Width, p.Height)
										//opts["vf"] = "scale=iw/2:ih/2"
										log.Println("Change bitrate", opts["b"])
									}
								}
							}
						} else {
							log.Println(err)
						}
						b.Edit(mm, "Transcoding ...")
						err = ffmpeg.Input(fname).
							Output(fname+".mp4", opts).OverWriteOutput().
							Run()
						if err != nil {
							log.Println(err)
						}
					} else {
						b.Edit(mm, "Probing...")
						f, err := ffmpeg.Probe(fname)
						var fjson FFProbe
						err = json.Unmarshal([]byte(f), &fjson)
						if err == nil {
							p.Width = int(fjson.Streams[0]["width"].(float64))
							p.Height = int(fjson.Streams[0]["height"].(float64))
						} else {
							log.Println(err)
						}
					}

					b.Edit(mm, "Sending...")

					p.File = tb.FromDisk(fname + ".mp4")
					p.FileName = "video.mp4"

					b.Send(admin, p)
					b.Delete(mm)
					if *datadirParam == "/tmp" {
						os.Remove(fname)
						os.Remove(fname + ".mp4")
					}
				} else {
					b.Send(admin, "Not found.")
				}
			} else if strings.HasPrefix(m.Text, "/video_") {
				args := strings.Split(m.Text[7:], "_")
				if len(args) > 1 {
					offset, err := strconv.Atoi(args[0])
					if err == nil {
						limit, err := strconv.Atoi(args[1])
						if err == nil && offset > -1 && limit > 0 {
							video(offset, limit)
						}
					}
				}
			}
		}
		done <- 1
	})

	go func() {
		for {
			i := <-motions
			if i.AlarmType == 3 {
				snapshot(false)
			} else {
				log.Println(i)
			}
		}
	}()

	b.Send(admin, "Bot restart!")
	b.Start()
}

func Login() int {
	// user = C.HLogin(C.CString(*ipParam), C.CString(*userParam), C.CString(*passParam), &dev)
	user, dev = hiklib.HikLogin(*ipParam, *userParam, *passParam)
	if int(user) > -1 {
		if *x1Param {
			hiklib.HikOnAlarmV30(user, *alarmParam, func(item hiklib.AlarmItem) {
				motions <- item
			})
		} else {
			hiklib.HikOnAlarm(user, *alarmParam, func(item hiklib.AlarmItem) {
				motions <- item
			})
		}
		return int(user)
	} else {
		return int(user)
	}
}

func main() {
	log.Println("HIKBOT v0.0.4")
	flag.Parse()
	if *ipParam == "" || *userParam == "" || *passParam == "" || *adminParam == 0 || *tkeyParam == "" {
		flag.PrintDefaults()
	} else {
		motions = make(chan hiklib.AlarmItem, 100)

		log.Printf("%s\n", hiklib.HikVersion())
		if Login() > -1 {
			defer hiklib.HikLogout(user)

			bot()
		} else {
			log.Println("Error login.")
		}
	}
}
