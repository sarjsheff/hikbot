#!/bin/sh

HIKSDKPATH=/hdd/EN-HCNetSDKV6.1.6.3_build20200925_Linux64
HIKUTILPATH=/hdd/hikutil

CGO_CFLAGS="-I$HIKUTILPATH" CGO_LDFLAGS="-L$HIKSDKPATH/lib -lhcnetsdk" go build

