.PHONY: build
build:
	go build -compiler gccgo -gccgoflags "-lX11" shift-shift.go
