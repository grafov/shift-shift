## What is it and why?

*Do you want to use Shift/Ctrl/Alt as modifiers while also switching language
layouts? This tool is for you then.*

A language layout switcher for Xorg and Wayland (including Sway and Hyprland)
that utilizes modifier keys (such as Shift and Control) as layout switchers.
This setup enables the same key to function both as a modifier and a layout
switcher without conflicts.

The utility incorporates two key concepts:

1. Use a key as a modifier when held and as a layout switcher when tapped.
2. Cyclic switching with a single key is inefficient; dedicated keys for each
   language group are more effective.

This approach may not be necessary when using multiple languages simultaneously,
but it is beneficial for most cases involving 2-3 languages. Frequent keyboard
layout switching, common among non-English speakers, is easier with dedicated
keys than with key combinations. For example, old Soviet computers featured a
dedicated RUS/LAT key to toggle between Latin and Cyrillic. Some computers even
offered two separate keys: RUS and LAT!

Unfortunately, modern English-oriented keyboards lack dedicated layout-switching
keys. Modifier keys are typically used only in combination with other keys. This
utility provides a solution: these keys can perform their original functions
while also being used to switch layouts.

By default, the Left Shift key switches to layout 1 and the Right Shift key to
layout 2. You can customize this behavior using command-line options. The utility
supports up to 4 groups.

Additionally, you can treat any input device as a keyboard using the `-match`
option, enabling simultaneous layout switching across multiple keyboards. The
utility listens to all connected devices.

The program listens to keypress events from the kernel (via evdev) and sends
them to Xorg or your window manager. Different WMs have their own methods for
keyboard switching, especially on Wayland. I have added support for Wayland
window managers I have used extensively. See the examples for all variants.

## Install

### Go way

This is a Go program. You need a Go environment to build it from source.

	go get github.com/grafov/shift-shift@latest

`xlib-devel` lib should be installed. Check your distro.

### Get sources

1. Get repository and cd to it.
2. `make build`

Prerequisites are the same: `xlib-devel` lib should be installed. Check your distro.

## Usage

```
$ Usage of ./shift-shift:
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
  -list-hypr
    	list all keyboards recognized by Hyprland
  -list-sway
    	list all devices recognized by Sway (not only keyboards)
  -match string
    	regexp used to match input keyboard device as it listed by evdev (default "keyboard")
  -match-wm string
    	optional regexp used to match device in WM, if not set evdev regexp used (default "keyboard")
  -print
    	print pressed keys for debug (verbose output)
  -quiet
    	be silent
  -scan-once
    	scan for keyboards only at startup (less power consumption)
  -switcher string
    	select method of switching (possible values are "auto", "xkb", "sway", "hypr") (default "auto")
```

## Configuration

### 1. Match keyboard in evdev output
First step is common for all environments. Just try to find your keyboard device. It may look like this:

```
$ shift-shift -list

/dev/input/event3 PC Speaker
/dev/input/event4 HDA ATI HDMI HDMI/DP,pcm=3
/dev/input/event6 HDA ATI HDMI HDMI/DP,pcm=8
/dev/input/event8 HD-Audio Generic Rear Mic
/dev/input/event12 HD-Audio Generic Front Headphone
/dev/input/event20 ZSA Technology Labs ErgoDox EZ System Control
/dev/input/event21 ZSA Technology Labs ErgoDox EZ Consumer Control
/dev/input/event23 VIRPIL Controls/20210102 R-VPC Stick MT-50
/dev/input/event7 HDA ATI HDMI HDMI/DP,pcm=9
/dev/input/event18 ZSA Technology Labs ErgoDox EZ
/dev/input/event22 ZSA Technology Labs ErgoDox EZ Keyboard
/dev/input/event24 ATMEL/VIRPIL/190930 BRD Rudder V3
/dev/input/event5 HDA ATI HDMI HDMI/DP,pcm=7
/dev/input/event17 Kensington      Kensington Expert Mouse
/dev/input/event19 ZSA Technology Labs ErgoDox EZ
...
```

And consists of any devices recognized by kernel on your machine. Just select any that looks like a keyboard :)

**Note:** you need setup proper access for reading `/dev/input/*`
devices. As a fallback try to run with `sudo` or similar tool.

You should add it `-match` key. You are free to use regular expression or just a substring for matching. Check [The
syntax of regexps](https://pkg.go.dev/regexp/syntax) for reference.

```
-match="^ZSA.*Keyboard$"
```

Hint. You could match several keyboards at once. They will used when connected:

```
$ shift-shift -match "Ergodox|Ergohaven"
```

Well, we've matched device with evdev. 

### 2. Match switcher

Next step is match it to a switcher. It
depends of your environment, see variants below.

#### 2.1 WM or DE under XOrg

```
$ shift-shift -match="^ZSA.*Keyboard$" -switcher=xkb
```

And it will run switcher especially for XOrg.

#### 2.2 Sway WM

In most of cases you should just add "-switcher=sway" to command line and it will works.

``` shell
$ shift-shift -match="^ZSA.*Keyboard$" -switcher=sway
```

You could list keyboards as they seen by Sway:

``` shell
$ shift-shift -list-sway

  {
    "identifier": "12951:18804:ZSA_Technology_Labs_ErgoDox_EZ_Keyboard",
    "name": "ZSA Technology Labs ErgoDox EZ Keyboard",
    "type": "keyboard",
    "repeat_delay": 350,
    "repeat_rate": 50,
    "xkb_layout_names": [
      "English (US)",
      "Russian (typewriter)"
    ],
    "xkb_active_layout_index": 0,
    "xkb_active_layout_name": "English (US)",
    "libinput": {
      "send_events": "enabled"
    },
    "vendor": 12951,
    "product": 18804
  },
  {
    "identifier": "12951:18804:ZSA_Technology_Labs_ErgoDox_EZ_Consumer_Control",
    "name": "ZSA Technology Labs ErgoDox EZ Consumer Control",
...
```

It will display your list of devices (but not only keyboards). 

Sometimes Sway may name the keyboard differently. In such cases, you can match
the keyboard using a regular expression similar to the evdev example above.
Then, add the result to the "-match-wm" key:

```
$ shift-shift -match="^ZSA.*Keyboard$" -switcher=sway -match-wm="^ZSA.*Keyboard$"
```

When "-match-wm" omited its value set to value of "-match".

#### 2.3 Hyprland WM

In most of cases you should just add "-switcher=hypr" to command line and it will works.

``` shell
$ shift-shift -match="(?i)ergodox" -switcher=hypr
```

You could list keyboards as they seen by Hyprland:

``` shell
$ shift-shift -list-hypr

model: name:zsa-technology-labs-ergodox-ez layout:us,ru options:caps:none,misc:typo,lv3:capslock_switch keymap:English (US)
```

Naming in Hyprland differs from evdev. You can match the keyboard using a
regular expression, similar to evdev as described above. Then, add the result to
the "-match-wm" key:

``` shell
$ shift-shift -match="^ZSA.*Keyboard$" -switcher=hypr -match-wm="^zsa-technology-labs-ergodox-ez$"
```

When "-match-wm" omited its value set to value of "-match".

#### 3. Print for detected keyboards

For debugging you could run with "-print" option. For example the output for Hyprland config above:

```
2025/01/30 01:12:28 use Hyprland switcher
2025/01/30 01:12:28 Hyprland keyboard matched zsa-technology-labs-ergodox-ez
2025/01/30 01:12:29 evdev keyboard found at /dev/input/event21: ZSA Technology Labs ErgoDox EZ Consumer Control
2025/01/30 01:12:29 evdev keyboard found at /dev/input/event22: ZSA Technology Labs ErgoDox EZ Keyboard

```

#### 4. Keys for switching

By default, the left Shift key is used for layout 0, and the right Shift key for
layout 1. You can use up to four layouts, naming them with the keys "-1" through
"-4". In the example below, the Ctrls key is used:

``` shell
$ shift-shift -match="^ZSA.*Keyboard$" -switcher=hypr -match-wm="^zsa-technology-labs-ergodox-ez$" -1 LEFT_CTRL -2 RIGHT_CTRL -print
```

The output with "-print" when you press any of switch keys:

```
2025/01/30 01:12:29 ZSA Technology Labs ErgoDox EZ type:1 code:90 pressed
2025/01/30 01:12:29 ZSA Technology Labs ErgoDox EZ type:1 code:90 released
2025/01/30 01:12:29 ZSA Technology Labs ErgoDox EZ switches group to 1
2025/01/30 01:12:29 switch hyprland kbd "zsa-technology-labs-ergodox-ez" to group 0
2025/01/30 01:12:30 ZSA Technology Labs ErgoDox EZ type:1 code:91 pressed
2025/01/30 01:12:30 ZSA Technology Labs ErgoDox EZ type:1 code:91 released
2025/01/30 01:12:30 ZSA Technology Labs ErgoDox EZ switches group to 2
```

For reference full list of key names recognized by evdev:

``` shell
$ cat /usr/include/linux/input-event-codes.h
```

#### 5. Final setup

Just remove "-print" once everything is working. Then, add the final command to
your window manager's autostart.

$ shift-shift -match="^ZSA.*Keyboard$" -switcher=hypr -match-wm="^zsa-technology-labs-ergodox-ez$" -1 LEFT_CTRL -2 RIGHT_CTRL -print

## Thanks

Thanks to people who contributed bugreports and improvements for
`shift-shift`, especially to
[@kovetsiy](https://github.com/kovetskiy),
[@ArtemT](https://github.com/ArtemT),
[@seletskiy](https://github.com/seletskiy).

Idea of Sway integration was inspired by Python code of
https://github.com/nmukhachev/sway-xkb-switcher project.
