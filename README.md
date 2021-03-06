wattson
=======

Wattson control library.
Written to pull generation/use data out of a "Wattson Solar Plus", but may work with a non-solar version.

Some of this is inspired from [openwattson](https://github.com/sapg/openwattson/blob/master/protocol.txt).

Which is in turn, inspired by [Mikko Pikarinen](http://dialog.hut.fi/openwattson/).
Thanks!

Protocol documentation
----------------------

Wattson talks over a USB interface that on most -nix machines will just show up at `/dev/ttyUSBx`.
It may require some stty wrangling, although this is handled by the library.

You send it commands followed by '\r\n' and it replies in kind.
Most replies are in the form "cNN[NN]", where 'c' is the command you sent, and 'NN' is a hex-encoded value (either one or two bytes).
Some commands may brick or reset your device, this is not an extensive list.

* nowp: Return the current power usage.
        This needs to be multiplied by the response to nown (plus one), and was likely done this way to suport usage over 65536 watts.

* nown: The multiplication factor to apply to nowp (plus one).
        If this is three, then multiply nowp by four to get the current usage.

* noww: The current generation.
        This value does *not* need to be muliplied by the response to nown.

* nowd: Count of days of power use stored, not including today. The response is
        in the form "dNN" where NN is the stored days count, in hex.

* nowl: This requests segment data for power use. Each day is divided into 12
        two-hour segments (each segment has 24 values of 5-minute periods).
        To request segment x on day y, request 'nowlyyxx'. These are 1-indexed,
        so the lowest day is 1, and the segments are in range 1-12.

  Note that the day and segment values are in dec, not hex (hex migh
  work - e.g., 1F will be treated the same as 25, which vaguely makes
  sense but is confusing while debugging).

  Finally, the 'current' period will always be indicated by FFFE (65534)
  and future periods will be indicated by FFFF (65535). If these values
  are not seen, then you could be looking at the wrong day.

* nowx: As per 'nowd', but for the number of days of generation stored. This is
        oddly sometimes a different number than 'nowd'.

  Note that this typically contains _only_ one or two days of data in the
  same form as 'nowd', e.g., this may return 1 and then there is data
  in day 1 and day 2 (being current).

* nowq: This contains a higher number representing the starting point of stored
        generation data that comes _after_ 'nowx'. e.g., 'nowx' may return 0,
        giving the first day in 1. 'nowq' might return 20, which seems to
        indicate following data is around here.

  Some testing:  
  +2 is two days ago  
  +3 is one day ago  
  +4 is crash/no data :-)  

* nowh: As per 'nowl', but for generation.

stty wrangling
--------------

The Wattson expects a very specific serial connection. This library now supports this (see [`tty.go`](lib/tty.go)), but previously this was set up manually via this magic incantation-

    /bin/stty -F /dev/ttyUSB1 19200 ignbrk -brkint -icrnl -imaxbel -opost -onlcr -isig -icanon -iexten -echo -echoe -echok noflsh -echoctl -echoke

