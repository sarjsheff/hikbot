#!/bin/sh

HIKSDKPATH=/hiksdk

CGO_CXXFLAGS="-I$HIKSDKPATH/incEn/" CGO_LDFLAGS="-L$HIKSDKPATH/lib -lhcnetsdk" go build
#CGO_CXXFLAGS="-I$HIKSDKPATH/incEn/" CGO_LDFLAGS="-L$HIKSDKPATH/lib -lhcnetsdk" go install
