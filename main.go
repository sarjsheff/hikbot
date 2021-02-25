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
// static void OnAlarm(int user) {
//   HListenAlarm(user,onmessage);
// }
// static void myprint(char* s) {
//   printf("%s\n", s);
// }
import "C"
import (
	"flag"
	"fmt"
	"log"
	"time"
	"unsafe"

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

type AlarmItem struct {
	IP        string
	Command   int
	AlarmType int
}

var motions chan AlarmItem

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

func bot(user C.int, byStartChan C.int) {
	var admin *tb.User
	b, err := tb.NewBot(tb.Settings{
		Token:  *tkeyParam,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/hello", func(m *tb.Message) {
		if m.Sender.ID == *adminParam {
			admin = m.Sender
			b.Send(admin, "Registered!")
		}
	})

	go func() {
		for {
			i := <-motions
			if i.AlarmType == 3 {
				fname := fmt.Sprintf("/tmp/%s.jpeg", i.IP)
				C.HCaptureImage(user, byStartChan, C.CString(fname))
				if admin != nil {
					p := &tb.Photo{File: tb.FromDisk(fname)}
					b.SendAlbum(admin, tb.Album{p})
				}
			} else {
				log.Println(i)
			}
		}
	}()

	b.Start()
}

func main() {
	flag.Parse()
	if *ipParam == "" || *userParam == "" || *passParam == "" || *adminParam == 0 || *tkeyParam == "" {
		flag.PrintDefaults()
	} else {
		motions = make(chan AlarmItem, 100)

		cs := C.CString("")
		dev := C.DevInfo{byStartChan: 0}
		C.HVersion(cs)
		log.Printf("%s\n", C.GoString(cs))
		user := C.HLogin(C.CString(*ipParam), C.CString(*userParam), C.CString(*passParam), &dev)
		if int(user) > -1 {
			C.OnAlarm(user)
			defer C.HLogout(user)
			bot(user, dev.byStartChan)
		} else {
			log.Println("Error login.")
		}
		C.free(unsafe.Pointer(cs))
	}
}
