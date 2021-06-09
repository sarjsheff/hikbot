package main

//// #cgo LDFLAGS: /hdd/hikutil/hiklib.o
// #include <stdio.h>
// #include <stdlib.h>
// #include "hiklib.h"
// #include "hik.h"
// #include <string.h>
//
// extern int onmessage(int command, char *sDVRIP, char *pBuf, unsigned int dwBufLen);
//
// static NET_DVR_ALARMINFO_V30 getalarminfo(char *pAlarmInfo) {
//    NET_DVR_ALARMINFO_V30 struAlarmInfo;
//    memcpy(&struAlarmInfo, pAlarmInfo, sizeof(NET_DVR_ALARMINFO_V30));
//    return struAlarmInfo;
// }
//
// static void OnAlarm(int user,int alarmport) {
//   printf("oalarm\n");
//   HListenAlarm(user,alarmport,onmessage);
// }
// static void myprint(char* s) {
//   printf("%s\n", s);
// }
import "C"
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
	"unsafe"

	"github.com/google/uuid"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

const COMM_ALARM int = 0x1100              //8000 Upload alarm message
const COMM_ALARM_V30 int = 0x4000          //9000 upload alarm message
const COMM_DEV_STATUS_CHANGED int = 0x7000 //Device status change alarm upload

var ipParam = flag.String("c", "", "Camera IP address.")
var userParam = flag.String("u", "", "Username.")
var passParam = flag.String("p", "", "Password.")
var tkeyParam = flag.String("t", "", "Telegram key.")
var adminParam = flag.Int("a", 0, "Telegram userid.")
var alarmParam = flag.Int("b", 7200, "Alarm port.")
var datadirParam = flag.String("d", "/tmp", "Data dir, default: /tmp .")
var previewsizeParam = flag.Int("s", 20000000, "Video preview byte size.")
var zParam = flag.Int("z", 2, "Video preview rescale (divide).")

type AlarmItem struct {
	IP        string
	Command   int
	AlarmType int
}

type FFProbe struct {
	Streams []map[string]interface{} `json:"streams"`
	Format  map[string]interface{}   `json:"format"`
}

var appid uuid.UUID = uuid.Must(uuid.NewRandom())

var motions chan AlarmItem
var dev = C.DevInfo{byStartChan: 0}
var user = C.int(-1)

type touser string
type MotionArea struct{ x, y, w, h float32 }

func (t touser) Recipient() string {
	return string(t)
}

//export onmessage
func onmessage(command C.int, ip *C.char, data *C.char, ln C.uint) C.int {

	i := AlarmItem{IP: C.GoString(ip), Command: int(command)}

	switch int(command) {
	case COMM_ALARM_V30:
		i.AlarmType = int(C.getalarminfo(data).dwAlarmType)
		motions <- i
		break
	case COMM_DEV_STATUS_CHANGED:
		log.Printf("COMM_DEV_STATUS_CHANGED %s %s", C.GoString(ip), C.GoString(data))
		break
	default:
		log.Printf("Unknown Alarm [0x%x] %s %s !!!", command, C.GoString(ip), C.GoString(data))
	}
	return 1
}

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
	video := func() {
		var v = C.MotionVideos{}
		C.HListVideo(user, &v)

		txt := ""
		for i := 0; i < int(v.count); i++ {
			dt := time.Date(int(v.videos[i].from_year), time.Month(int(v.videos[i].from_month)), int(v.videos[i].from_day), int(v.videos[i].from_hour), int(v.videos[i].from_min), int(v.videos[i].from_sec), 0, time.UTC)
			todt := time.Date(int(v.videos[i].to_year), time.Month(int(v.videos[i].to_month)), int(v.videos[i].to_day), int(v.videos[i].to_hour), int(v.videos[i].to_min), int(v.videos[i].to_sec), 0, time.UTC)
			txt = txt + "<b>" + dt.Format("2006-01-02 15:04:05") + " - " + todt.Format("15:04:05") + "</b> /dl_" + C.GoString(v.videos[i].filename) + " \n"
			videolist[C.GoString(v.videos[i].filename)] = dt.Format("2006-01-02/15:04:05")
		}
		// menu.Inline(menu.Row(btnSettings))
		// b.Send(admin, txt, &tb.SendOptions{ReplyMarkup: menu, ParseMode: tb.ModeHTML})
		b.Send(admin, txt, &tb.SendOptions{ParseMode: tb.ModeHTML})
	}

	snapshot := func(mareas bool) {
		fname := filepath.Join(*datadirParam, fmt.Sprintf("%s.jpeg", uuid.Must(uuid.NewRandom()).String()))
		err := C.HCaptureImage(user, dev.byStartChan, C.CString(fname))
		if err > -1 {
			caption := ""
			if mareas {
				var ma = C.MotionAreas{}
				C.HMotionArea(user, &ma)

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
					if ma.areas[i].w > 0 && ma.areas[i].h > 0 {
						x, y, w, h := int(float32(b.Dx())*float32(ma.areas[i].x)), int(float32(b.Dy())*float32(ma.areas[i].y)), int(float32(b.Dx())*float32(ma.areas[i].w)), int(float32(b.Dy())*float32(ma.areas[i].h))
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
			video()
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
			res := C.HReboot(user)
			if int(res) > 0 {
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
						C.HSaveFile(user, C.CString(m.Text[4:]), C.CString(fname))
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
			}
		}
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
	user = C.HLogin(C.CString(*ipParam), C.CString(*userParam), C.CString(*passParam), &dev)
	if int(user) > -1 {
		C.OnAlarm(user, C.int(*alarmParam))
		return int(user)
	} else {
		return int(user)
	}
}

func main() {
	log.Println("HIKBOT " + appid.String())
	flag.Parse()
	if *ipParam == "" || *userParam == "" || *passParam == "" || *adminParam == 0 || *tkeyParam == "" {
		flag.PrintDefaults()
	} else {
		motions = make(chan AlarmItem, 100)

		cs := C.CString("")

		C.HVersion(cs)
		log.Printf("%s\n", C.GoString(cs))
		if Login() > -1 {
			defer C.HLogout(user)

			bot()
		} else {
			log.Println("Error login.")
		}
		C.free(unsafe.Pointer(cs))
	}
}
