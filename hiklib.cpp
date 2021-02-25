#include <stdio.h>

#include <cstring>

#include "HCNetSDK.h"
#include "hiklib.h"

unsigned int HVersion(char *ret)
{
  NET_DVR_Init();
  unsigned int uiVersion = NET_DVR_GetSDKBuildVersion();
  sprintf(ret, "HCNetSDK V%d.%d.%d.%d", (0xff000000 & uiVersion) >> 24, (0x00ff0000 & uiVersion) >> 16, (0x0000ff00 & uiVersion) >> 8, (0x000000ff & uiVersion));
  return uiVersion;
}

int HLogin(char *ip, char *username, char *password, struct DevInfo *devinfo)
{
  char logPath[11] = "./sdkLog.v";
  NET_DVR_Init();
  NET_DVR_SetLogToFile(0, logPath);
  NET_DVR_USER_LOGIN_INFO struLoginInfo = {0};
  NET_DVR_DEVICEINFO_V40 struDeviceInfoV40 = {0};
  struLoginInfo.bUseAsynLogin = false;

  struLoginInfo.wPort = 8000;
  memcpy(struLoginInfo.sDeviceAddress, ip, NET_DVR_DEV_ADDRESS_MAX_LEN);
  memcpy(struLoginInfo.sUserName, username, NAME_LEN);
  memcpy(struLoginInfo.sPassword, password, NAME_LEN);

  int lUserID = NET_DVR_Login_V40(&struLoginInfo, &struDeviceInfoV40);

  if (lUserID < 0) {
    int err = NET_DVR_GetLastError();
    printf("\n\nError %d\n\n", err);
    NET_DVR_Cleanup();
    return 0 - err;
  }

  devinfo->byStartChan = struDeviceInfoV40.struDeviceV30.byStartChan;

  return lUserID;
}

void HLogout(int lUserID)
{
  NET_DVR_Logout_V30(lUserID);
  NET_DVR_Cleanup();
}

int HCaptureImage(int lUserID, int byStartChan, char *imagePath)
{
  printf("Capture image to [%s].\n", imagePath);
  NET_DVR_JPEGPARA strPicPara = {0};
  strPicPara.wPicQuality = 0;
  strPicPara.wPicSize = 0xff;
  int iRet;
  iRet = NET_DVR_CaptureJPEGPicture(lUserID, byStartChan, &strPicPara, imagePath);
  if (!iRet) {
    return NET_DVR_GetLastError();
  }
  return iRet;
}

BOOL CALLBACK OnMessage(int lCommand, char *sDVRIP, char *pBuf, DWORD dwBufLen)
{
  printf("OnMessage %d from %s [%s]%d\n", lCommand, sDVRIP, pBuf, dwBufLen);
  return true;
}

int HListenAlarm(long lUserID,
                 int (*fMessCallBack)(int lCommand,
                                      char *sDVRIP,
                                      char *pBuf,
                                      unsigned int dwBufLen))  // BOOL(CALLBACK *fMessCallBack)(LONG lCommand, char *sDVRIP, char *pBuf, DWORD dwBufLen))
{
  NET_DVR_SetDVRMessCallBack(fMessCallBack);
  if (NET_DVR_StartListen(NULL, 7200)) {
    printf("Start listen\n");
    LONG m_alarmHandle = NET_DVR_SetupAlarmChan(lUserID);
    if (m_alarmHandle < -1) {
      return 0 - NET_DVR_GetLastError();
    } else {
      return m_alarmHandle;
    }
  } else {
    printf("\n\nError start alarmlisten\n\n");
    return -1;
  }
}

