// trmnl-input creates a short-lived virtual touchscreen for on-device UI tests.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

const (
	uiSetEvBit      = 0x40045564
	uiSetKeyBit     = 0x40045565
	uiSetAbsBit     = 0x40045567
	uiSetPropBit    = 0x4004556e
	uiDevSetup      = 0x405c5503
	uiAbsSetup      = 0x401c5504
	uiDevCreate     = 0x5501
	uiDevDestroy    = 0x5502
	evSyn           = 0
	evKey           = 1
	evAbs           = 3
	synReport       = 0
	btnTouch        = 330
	absMTSlot       = 47
	absMTTouchMajor = 48
	absMTPositionX  = 53
	absMTPositionY  = 54
	absMTToolType   = 55
	absMTTrackingID = 57
	absMTPressure   = 58
	absMTDistance   = 59
	inputPropDirect = 1
)

type inputID struct{ BusType, Vendor, Product, Version uint16 }
type deviceSetup struct {
	Name         [80]byte
	ID           inputID
	FFEffectsMax uint32
}
type absInfo struct{ Value, Minimum, Maximum, Fuzz, Flat, Resolution int32 }
type absSetup struct {
	Code uint16
	Pad  uint16
	Info absInfo
}

func ioctl(fd uintptr, request uintptr, arg uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, request, arg)
	if e != 0 {
		return e
	}
	return nil
}
func setBit(fd uintptr, request uintptr, value int) error { return ioctl(fd, request, uintptr(value)) }

func main() {
	if len(os.Args) < 4 {
		fatal("usage: trmnl-input tap X Y [hold_ms] | swipe X1 Y1 X2 Y2 duration_ms")
	}
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		fatal(err.Error())
	}
	defer f.Close()
	fd := f.Fd()
	for _, v := range []struct {
		req uintptr
		bit int
	}{{uiSetEvBit, evSyn}, {uiSetEvBit, evKey}, {uiSetEvBit, evAbs}, {uiSetKeyBit, btnTouch}, {uiSetPropBit, inputPropDirect}} {
		if err := setBit(fd, v.req, v.bit); err != nil {
			fatal(err.Error())
		}
	}
	abs := map[int]absInfo{absMTSlot: {Minimum: 0, Maximum: 9}, absMTTouchMajor: {Minimum: 0, Maximum: 255}, absMTPositionX: {Minimum: 0, Maximum: 2064, Resolution: 2064}, absMTPositionY: {Minimum: 0, Maximum: 2832, Resolution: 2832}, absMTToolType: {Minimum: 0, Maximum: 2}, absMTTrackingID: {Minimum: 0, Maximum: 65535}, absMTPressure: {Minimum: 0, Maximum: 255}, absMTDistance: {Minimum: 0, Maximum: 255}}
	for code, info := range abs {
		if err := setBit(fd, uiSetAbsBit, code); err != nil {
			fatal(err.Error())
		}
		s := absSetup{Code: uint16(code), Info: info}
		if err := ioctl(fd, uiAbsSetup, uintptr(unsafe.Pointer(&s))); err != nil {
			fatal(err.Error())
		}
	}
	var setup deviceSetup
	copy(setup.Name[:], "TRMNL integration touch")
	setup.ID = inputID{BusType: 0x06, Vendor: 0x5452, Product: 0x4d4e, Version: 1}
	if err := ioctl(fd, uiDevSetup, uintptr(unsafe.Pointer(&setup))); err != nil {
		fatal(err.Error())
	}
	if err := ioctl(fd, uiDevCreate, 0); err != nil {
		fatal(err.Error())
	}
	defer ioctl(fd, uiDevDestroy, 0)
	time.Sleep(1500 * time.Millisecond)
	switch os.Args[1] {
	case "tap":
		x := number(2)
		y := number(3)
		hold := 150
		if len(os.Args) > 4 {
			hold = number(4)
		}
		touch(f, x, y, time.Duration(hold)*time.Millisecond)
	case "swipe":
		if len(os.Args) != 7 {
			fatal("swipe requires X1 Y1 X2 Y2 duration_ms")
		}
		swipe(f, number(2), number(3), number(4), number(5), time.Duration(number(6))*time.Millisecond)
	default:
		fatal("unknown action")
	}
	time.Sleep(800 * time.Millisecond)
}

func touch(f *os.File, x, y int, hold time.Duration) { down(f, x, y, 1); time.Sleep(hold); up(f) }
func swipe(f *os.File, x1, y1, x2, y2 int, d time.Duration) {
	steps := 20
	down(f, x1, y1, 2)
	for i := 1; i <= steps; i++ {
		time.Sleep(d / time.Duration(steps))
		emit(f, evAbs, absMTPositionX, x1+(x2-x1)*i/steps)
		emit(f, evAbs, absMTPositionY, y1+(y2-y1)*i/steps)
		syncEvent(f)
	}
	up(f)
}
func down(f *os.File, x, y, id int) {
	emit(f, evAbs, absMTSlot, 0)
	emit(f, evAbs, absMTTrackingID, id)
	emit(f, evAbs, absMTPositionX, x)
	emit(f, evAbs, absMTPositionY, y)
	emit(f, evAbs, absMTTouchMajor, 30)
	emit(f, evAbs, absMTPressure, 80)
	emit(f, evKey, btnTouch, 1)
	syncEvent(f)
}
func up(f *os.File) {
	emit(f, evAbs, absMTSlot, 0)
	emit(f, evAbs, absMTTrackingID, -1)
	emit(f, evKey, btnTouch, 0)
	syncEvent(f)
}
func syncEvent(f *os.File) { emit(f, evSyn, synReport, 0) }
func emit(f *os.File, typ, code, value int) {
	b := make([]byte, 24)
	binary.LittleEndian.PutUint16(b[16:18], uint16(typ))
	binary.LittleEndian.PutUint16(b[18:20], uint16(code))
	binary.LittleEndian.PutUint32(b[20:24], uint32(int32(value)))
	if _, err := f.Write(b); err != nil {
		fatal(err.Error())
	}
}
func number(i int) int {
	n, e := strconv.Atoi(os.Args[i])
	if e != nil {
		fatal(e.Error())
	}
	return n
}
func fatal(s string) { fmt.Fprintln(os.Stderr, s); os.Exit(2) }
