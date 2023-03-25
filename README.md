What is it and why?
===================

Language layout switcher for Xorg and Wayland (Sway support) with
modifiers (Shift, Control and so on) as switcher keys.

This way allows to use the same key as a modifier and as a layout switcher without conflicts.

The utility implements two ideas:

1. You could use a key as a modifier when you HOLD it and as a key switcher when you TAP it.
2. Cyclic switching by a single key is a bad idea, better to use DEDICATED keys for each language group.

It maybe not true if you use a bunch of languages of same time but it
is true for the most use cases with 2-3 langs.

If you are often switch keyboard layouts (it real use case
for those who speaking not only English) then dedicated keys are more easy
for typing than key combos. Old Soviet computers for example had
dedicated key RUS/LAT for switch between Latin and Cyrillic.

Sadly in modern English-oriented keyboards there are no dedicated keys
for switching layouts. Modifier keys usually used only by holding with
other key. So it looks like a good compromise: you are still able to
use them for their original purposes but when you tapping them they
work as language layout switchers.

By default Left Shift swithches to group 1 and Right Shift to
group 2. You may change this behavior with command line options.
Up to 4 xkb groups supported.

Also you could try to treat any devices as keyboards with `-match`
option. It allows to switch group simultaneously on an arbitrary
number of connected keyboards.

Install
=======

This is a Go program. You should need Go environment to build it from sources.

	go get github.com/grafov/shift-shift@latest

`xlib-devel` libs should be installed. Check your distro.

Usage
=====

```
$ shift-shift -h
Usage of shift-shift:
  -1 string
		key used for switching to 1st xkb group (default "LEFTSHIFT")
  -2 string
		key used for switching to 2nd xkb group (default "RIGHTSHIFT")
  -3 string
		key used for switching to 3rd xkb group
  -4 string
		key used for switching to 4th xkb group
  -double-keystroke
		require pressing the same key twice to switch the layout
  -double-keystroke-timeout int
		second keystroke timeout in milliseconds (default 500)
  -list
		list all devices that found by evdev (not only keyboards)
  -list-sway
		list all devices recognized by Sway (not only keyboards)
  -match string
		regexp used to match keyboard device (default "keyboard")
  -print
		print pressed keys for debug (verbose output)
  -quiet
		be silent
  -scan-once
		scan for keyboards only at startup (less power consumption)
  -switcher string
		select method of switching (possible values are "auto", "xkb", "sway") (default "auto")
```

On start the program tries to find devices where name contains "keyboard" string. The substring could be customized with `-match` option. [The syntax](https://pkg.go.dev/regexp/syntax) of regular expressions could be used. For example:
```
$ shift-shift -match "(?i)ergodox|ergohaven"
```

Check the list of evdev detected devices with:

```
$ shift-shift -list
```

**Wayland/Sway note.** In the same time `-match` option applied to the
list of Sway input devices. You could check them with:

```
$ shift-shift -list-sway
```

**Note:** you need setup proper access for reading `/dev/input/*`
devices. As a fallback try to run with `sudo` or similar tool.

Thanks
======

Thanks to people who contributed bugreports and improvements for
`shift-shift`, especially to
[@kovetsiy](https://github.com/kovetskiy),
[@ArtemT](https://github.com/ArtemT),
[@seletskiy](https://github.com/seletskiy).

Idea of Sway integration was inspired by Python code of
https://github.com/nmukhachev/sway-xkb-switcher project.
