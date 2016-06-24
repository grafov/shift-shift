What is it and why?
===================

This utility allows you switch keyboard groups in X Window in the most
ergonomic way (in my view :). 
I think keys for switching keyboard layouts should be:

1. Dedicated.
2. Non modal.

If you are often switch keyboard layouts (it real use case
for those who speaking not only English) then dedicated keys are more easy
for typing than key combos.  Old Soviet computers for example had
dedicated key RUS/LAT for switch between Latin and Cyrillic. Sadly
in modern English-oriented keyboards there are no dedicated keys
for switch layouts.

So if you need two layouts you need two keys each of them will select
only one layout. For example if you want to type Cyrillic you press switch
key dedicated for this layout. Other dedicated key switches keyboard to Latin.
You may press switcher keys many times but each key will still select its own
layout. Such non modal way minimizes number of mistakes and allow you to work
without visual indication of current layout.

After experiments with different control keys I found that most
comfortable for my fingers will use LShift/RShift. LShift dedicated
for first layout and RShift for second when they pressed as standalone
keys. But when you press Shift with other keys then it applied as
modifier key.

I not found out of the box solution how to setup X to use Shifts as
standalone keys. Also I used two keyboards in same time (notebok internal and USB plugged) 
and was need to switch layouts on both of them. So I wrote this utility.

So `shift-shift` has features:

* LShift pressed standalone locks X to group1 layout
* RShift pressed standalone locks X to group2 layout
* Layout switched on all keyboards simultaneously

You need customize layout groups in your X config or with `setxkbmap`.

About code
==========

I used Go language as it allow to me write programs quickly. And it allow
easy to combine it with C code. Though whole program may be rewritten to C
but I lazy to do it as it already works as I need.

I not well know programming for X so used simple and crude ways to do things.
Program use udev library and requires root privileges for reading `/dev/input/event*`.
If you know right way how to do it without root then let me advice please.

Install
=======

Binding of `evdev` for Go used so before build you need:

    go get github.com/gvalkov/golang-evdev/evdev

Then as usual:

    go build

Of course you need Go environment installed for build. 
And as program uses Xlib through cgo interface then you need `xlib-devel`
installed.

Usage
=====

     $ sudo shift-shift -h
     Usage of shift-shift:
       -list=false: list all devices listened by evdev
       -match="keyboard": string used to match keyboard device
       -print=false: print pressed keys
       -quiet=false: be silent

On start program find devices where name contains "keyboard" string. It assume there
are keyboard devices. You may customize this by set your own string with `-match` arg.
Got list of all input devices with `list` arg.

**Note:** you need root for operations with `/dev/input/*` so run it with `sudo` or similar tool.

For autostart run it somewhere after X started with your account. I use `~/.bash_profile` for
this.

    sudo pidof shift-shift >/dev/null || sudo shift-shift -quiet >/dev/null &

Thanks
======

Thanks to people who contributed bugfixes and improvements for `shift-shift`.
