# Hikvision telegram bot

Listen alarm events form hikvision camera and take snapshot.

```
HIKUTILDIR=/hdd/hikutil/ LD_LIBRARY_PATH=/hdd/EN-HCNetSDKV6.1.6.3_build20200925_Linux64/lib ./hikbot -t "telegramkey" -u username -p password -c cameraip -a telegram user id
```

# Telegram commands

```/snap``` - take snapshot from camera.

```/reboot``` - reboot camera.

```/video``` - list saved video from camera with download command.

```/mareas``` - show motion areas.

[Motion events](imgs/event.png)

[Motion areas](imgs/mareas.png)

[Video list](imgs/video.png)
