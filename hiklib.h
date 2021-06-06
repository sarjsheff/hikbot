#ifdef __cplusplus
extern "C"
{
#endif

  typedef struct DevInfo
  {
    int byStartChan;
  } DevInfo;

  typedef struct MotionVideo
  {
    char *filename;
    long size;
    int from_year;
    int from_month;
    int from_day;
    int from_hour;
    int from_min;
    int from_sec;
    int to_year;
    int to_month;
    int to_day;
    int to_hour;
    int to_min;
    int to_sec;
  } MotionVideo;

  typedef struct MotionVideos
  {
    MotionVideo videos[100];
    int count;
  } MotionVideos;

  typedef struct MotionArea
  {
    float x;
    float y;
    float w;
    float h;
  } MotionArea;

  typedef struct MotionAreas
  {
    MotionArea areas[8];
  } MotionAreas;

  unsigned int HVersion(char *ret);
  int HLogin(char *ip, char *username, char *password, struct DevInfo *devinfo);
  void HLogout(int lUserID);
  int HMotionArea(int lUserID, struct MotionAreas *areas);
  int HCaptureImage(int lUserID, int byStartChan, char *imagePath);
  int HListenAlarm(long lUserID, int alarmport, int (*fMessCallBack)(int lCommand, char *sDVRIP, char *pBuf, unsigned int dwBufLen));
  int HReboot(int user);
  int HListVideo(int lUserID, struct MotionVideos *videos);
  int HSaveFile(int userId, char *srcfile, char *destfile);

#ifdef __cplusplus
}
#endif
