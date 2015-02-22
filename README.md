What is and why?
================

This utility switch keyboard groups in X Window in a way most
ergonomic on my taste. I think switch keyboard layouts must be:

1. By dedicated keys.
2. Non modal.

If you often switch different keyboard layouts (it real use case
for all who speak not only English) then dedicated keys more easy
to type than key combos. Old Soviet computers for example have
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
standalone keys. Also I used two keyboards in same time and was need
to switch layouts on both of them. So I wrote this utility.

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
If you know right way how to do it without root then write me.

Install
=======

Binding of `udev` for Go used so before build you need:

    go get github.com/gvalkov/golang-evdev/evdev

Then as usual:

    go build

Of course you need Go environment installed for build. 
And as program uses Xlib through cgo interface then you need `xlib-devel`
installed.

Run
====

Run somewhere after X started with your account. I use `~/.bash_profile` for
this.

    sudo pidof shift-shift >/dev/null || sudo shift-shift -quiet >/dev/null &

Thanks
======

Thanks to people who contributed bugfixes and improvements for `shift-shift`.
