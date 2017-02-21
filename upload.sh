set -ex
rm -f ./main
go build main.go
ssh  forth.yak.net  killall rxtx_main.bin || true
scp ./main forth.yak.net:/opt/disk/public/rxtx_main.bin
