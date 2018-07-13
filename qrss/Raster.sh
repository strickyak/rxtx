alias rig="rigctl -m 229 -r /dev/ttyUSB0 -s 38400"

while true
do
	# rig T 1

	go run 2>_ main.go --mode=raster -secs=1.0 --ramp=0.05 --base=1000 --step=3 |
	  pacat --format=s16be --channels=1 --channel-map=mono  --rate=44100 --device=alsa_output.usb-Burr-Brown_from_TI_USB_Audio_CODEC-00.analog-stereo  

	# rig T 0
	sleep 60
done
