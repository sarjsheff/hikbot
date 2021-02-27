#ifdef __cplusplus
extern "C" {
#endif

typedef struct DevInfo {
  int byStartChan;
} DevInfo;

unsigned int HVersion(char *ret);
int HLogin(char *ip, char *username, char *password, struct DevInfo *devinfo);
void HLogout(int lUserID);
int HCaptureImage(int lUserID, int byStartChan, char *imagePath);
int HListenAlarm(long lUserID, int alarmport, int (*fMessCallBack)(int lCommand, char *sDVRIP, char *pBuf, unsigned int dwBufLen));
int HReboot(int user);

#ifdef __cplusplus
}
#endif
