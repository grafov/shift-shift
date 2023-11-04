What is it and why?
===================

Language layout switcher for Xorg and Wayland (Sway & River WM are
supported).

The utility distinguishes when the user tap (press and release) a key
vs continuos holding of a key. It allows to use modifier keys (like
Shift key for example) as a layout switcher as modifier in the same
time.

The main design idea is:

> Cyclic switching by a single key is a bad idea. Dedicated keys should
> be used for switching. Especially when more then two layouts in use.

By default it uses Left Shift for group 1 and Right Shift for group 2
for xkb. This way makes special indicators for current layout just
useless. Because when you want to type sowewhat in the specific layout
you are just tap switcher key before. It is not a problem if you press
it again because there are no cycling switching. By pressing the key
you be sure *which* layout you are selected.

Up to 4 xkb groups supported for Xorg. For Wayland
[sway](https://swaywm.org/) and
[river](https://github.com/riverwm/river) window managers are
supported. Check the `-switcher` opion.

You could try to treat any devices as keyboards with `-match`
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
		list all devices recognized by Sway WM (not only keyboards)
  -match string
		regexp used to match keyboard device (default "keyboard")
  -print
		print pressed keys for debug (verbose output)
  -quiet
		be silent
  -scan-once
		scan for keyboards only at startup (less power consumption)
  -switcher string
		select method of switching (possible values are "auto", "xkb", "river", "sway") (default is "auto")
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
