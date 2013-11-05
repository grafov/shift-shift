shift-shift
===========

Simple Xorg keyboard switcher with non modal keyboard combination.
Firstly I want to get convinient tool for my needs that I hardcoded
switch key. Later I will customize it. Now simply:

* Left Shift — lock to X group1
* Right Shift — lock to X group2

It is feature: key used for switch non modal. You press Lshift and may press it repeatly - then
it will locked to group1 (English language in my X config) and press Rshift will lock to only
group2 (Russian language in my case).

Program requires root privileges for reading /dev/input/event*.

Build
-----

You need installed go language, then use standard for go-programs way:

				go build

It uses Xlib through cgo interface so you need installed xlib-devel library.
