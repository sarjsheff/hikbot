
#define MAX_ANALOG_ALARMOUT         32      //32 analog alarm output channels intotal
#define MAX_IP_ALARMOUT             64      //Maximum number of alarm output channels that can be added
#define MAX_ALARMOUT_V30               ( MAX_ANALOG_ALARMOUT + MAX_IP_ALARMOUT )
#define MAX_CHANNUM_V30             64
#define MAX_DISKNUM_V30                33

typedef  unsigned int       DWORD;
typedef  unsigned short     WORD;
typedef  unsigned short     USHORT;
typedef  short              SHORT;
typedef  int                LONG;
typedef  unsigned char      BYTE;
typedef  unsigned int       UINT;
typedef  void*              LPVOID;
typedef  void*              HANDLE;
typedef  unsigned int*      LPDWORD;
typedef  unsigned long long UINT64;
typedef  signed long long   INT64;


//Upload Alarm Information  (9000 extended)
typedef struct
{
    DWORD dwAlarmType; // 0- sensor alarm; 1- hard disk full; 2- video lost; 3- motion detection; 4- hard disk unformatted;
////                         5- hard disk error; 6- tampering detection; 7- unmatched video output standard; 8- illegal operation;
////                         9- video exception; 0xa- record exception
////                       11- Vca scene change 12-Array exception 13 resolution dismatch,14-alloc decode resource fail,
////                       15-VCA detect, 16-POE power supply abnormal alarm,17-Flash anomaly ,18-Disk full load anomaly,
////                       19-audio input lost, 20-record on, 21-record off,22-vehicle detection exception, 23-pulse alarm,
////                       24-face lib disk alarm,25-face lib change,26-face picture change,28-camera angle anomaly
////                       ,29-battery low,30-Lack of SD card
    DWORD dwAlarmInputNumber; // Alarm input Port
    BYTE byAlarmOutputNumber[MAX_ALARMOUT_V30]; // State of Alarm output channel, 1- - Triggered
    BYTE byAlarmRelateChannel[MAX_CHANNUM_V30]; //channels triggered to record, 1- recording,  dwAlarmRelateChannel[0] is the first channel
    BYTE byChannel[MAX_CHANNUM_V30]; //If the dwAlarmType is 2, 3 , 6,14, 19 or 28 it stands for channel, dwChannel[0] is the first channel
    BYTE byDiskNumber[MAX_DISKNUM_V30]; //When dwAlarmType is 1, 4 or 5,  it stands for Hard Disk,  dwDiskNumber[0] is the first disk
}NET_DVR_ALARMINFO_V30, *LPNET_DVR_ALARMINFO_V30;

