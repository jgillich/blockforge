default: cgo

cgo: xmr-stak

xmr-stak:
	cmake -Bcgo/xmr-stak -Hcgo/xmr-stak -DCUDA_ENABLE=OFF -DOpenCL_ENABLE=OFF
	$(MAKE) -C cgo/xmr-stak
