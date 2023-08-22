# Hikvision telegram bot

Listen alarm events form hikvision camera and take snapshot.

```
HIKUTILDIR=/hdd/hikutil/ LD_LIBRARY_PATH=/hdd/EN-HCNetSDKV6.1.6.3_build20200925_Linux64/lib ./hikbot -t "telegramkey" -u username -p password -c cameraip -a telegram user id
```

# Build

Dagger ci:

```
go run ci/ci.go -s "path to HCNetSDKV6.1.9.48_build20230410_linux64" -r "registry url" -i "publish image" -u 'registry user' -p "registry password"
```

# Telegram commands

```/info``` - camera info.

```/snap``` - take snapshot from camera.

```/reboot``` - reboot camera.

```/video``` - list saved video from camera with download command.

```/mareas``` - show motion areas.

![Motion events](imgs/event.png?raw=true)

![Motion areas](imgs/mareas.png?raw=true)

![Video list](imgs/video.png?raw=true)
