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
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"

	"github.com/google/uuid"
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

type AlarmItem struct {
	IP        string
	Command   int
	AlarmType int
}

var appid uuid.UUID = uuid.Must(uuid.NewRandom())

var motions chan AlarmItem
var dev = C.DevInfo{byStartChan: 0}
var user = C.int(-1)

type touser string

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

	snapshot := func() {
		fname := fmt.Sprintf("/tmp/%s.jpeg", uuid.Must(uuid.NewRandom()).String())
		err := C.HCaptureImage(user, dev.byStartChan, C.CString(fname))
		if err > -1 {
			p := &tb.Photo{File: tb.FromDisk(fname)}
			b.SendAlbum(admin, tb.Album{p})
			os.Remove(fname)
		} else {
			b.Send(admin, fmt.Sprintf("Error get snapshot [%d].", err))
		}
	}

	b.Handle("/snap", func(m *tb.Message) {
		<-done
		if m.Sender.ID == *adminParam {
			snapshot()
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

	go func() {
		for {
			i := <-motions
			if i.AlarmType == 3 {
				snapshot()
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
