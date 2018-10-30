module gitlab.com/gomidi/rtmididrv

replace gitlab.com/gomidi/rtmididrv/imported/rtmidi => ./imported/rtmidi

require (
	gitlab.com/gomidi/midi v1.7.4
	gitlab.com/gomidi/rtmididrv/imported/rtmidi v0.0.0-20181023173540-4751d32e0b95
)
