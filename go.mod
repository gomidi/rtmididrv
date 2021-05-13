module gitlab.com/gomidi/rtmididrv

go 1.14

require (
	gitlab.com/gomidi/midi v1.21.0
	gitlab.com/gomidi/rtmididrv/imported/rtmidi v0.11.0
)

replace (
	gitlab.com/gomidi/rtmididrv/imported/rtmidi => ./imported/rtmidi/
)
