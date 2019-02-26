module gitlab.com/gomidi/rtmididrv

replace gitlab.com/gomidi/rtmididrv/imported/rtmidi => ./imported/rtmidi

require (
	gitlab.com/gomidi/midi v1.10.0
	gitlab.com/gomidi/rtmididrv/imported/rtmidi v0.0.0-20181030132923-7607b12e13d8
)
