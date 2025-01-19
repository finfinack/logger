package logging

import (
	"fmt"
	"strings"
	"time"
)

const (
	// Log severities are used as prefixes and can max be 5 characters.
	logPrefixDebug   = "DEBUG"
	logPrefixInfo    = "INFO"
	logPrefixWarn    = "WARN"
	logPrefixError   = "ERROR"
	logPrefixFatal   = "FATAL"
	logPrefixUnknown = "N/A"
)

const (
	LogLevelDebug = 0
	LogLevelInfo  = 10
	LogLevelWarn  = 20
	LogLevelError = 30
	LogLevelFatal = 40
)

const (
	contextCheckDelay = time.Second
	logTimeFormat     = time.RFC3339
)

var logLevelToName = map[int]string{
	LogLevelDebug: logPrefixDebug,
	LogLevelInfo:  logPrefixInfo,
	LogLevelWarn:  logPrefixWarn,
	LogLevelError: logPrefixError,
	LogLevelFatal: logPrefixFatal,
}

func LevelToName(lvl int) (string, error) {
	n, ok := logLevelToName[lvl]
	if !ok {
		return "", fmt.Errorf("unable to find log level corresponding to %d", lvl)
	}
	return n, nil
}

func LevelToValue(name string) (int, error) {
	name = strings.ToUpper(name)
	for lvl, n := range logLevelToName {
		if n == name {
			return lvl, nil
		}
	}
	return -1, fmt.Errorf("unable to find log level corresponding to %q", name)
}

var logExitMessages = []string{
	"Das System macht Feierabend. Bitte nicht stören. 🛑",
	"Herunterfahren... und nein, es war nicht meine Schuld! 🤷‍♂️",
	"Das war's für heute. Tschüssikowski! 👋",
	"Ich beende mich. Bitte weine nicht. 😢",
	"Ich schliesse... aber die offenen Tabs? Deine Sache. 🤪",
	"Diese App ist offiziell offline. Prost! 🍻",
	"Ich geh' schlafen, weck mich nicht vor dem Update. 😴",
	"Die Anwendung ist weg, aber die Bugs bleiben. 🐞",
	"Herunterfahren, wie gewünscht. Zufrieden? 🫠",
	"Programm beendet. Das war der letzte Kaffee für heute. ☕",
	"Das war's, ich bin raus. Wer macht jetzt den Abwasch? 🍽️",
	"Alle Daten gesichert. Und nein, ich verrate nicht wohin. 🤐",
	"System offline. Jetzt gibt's erst mal Feierabendbier. 🍺",
	"Der Prozess ist tot, es lebe der Prozess! 🪦",
	"Shutting down... because this code needs its beauty sleep. 😴",
	"This is not a bug - it's a happy little accident. 🎨",
	"Closing... but I'll be back. Maybe. 💪",
	"This app is officially taking a coffee break. ☕",
	"Shutting down because someone hit the big red button. 🛑",
	"End of process. Did you save your work? 🤔",
	"This is your application signing off. Peace! ✌️",
	"Closing because someone forgot to feed the server. 🍔",
	"Goodbye for now. Don't miss me too much. 💔",
	"Exiting gracefully... like a cow on ice. 🐄",
	"Logging out. Don't forget to tip your developer. 💰",
	"All good things must come to an end... including me. 🌈",
	"I'm gone. But the bugs remain. 🐛",
	"Hasta la vista, baby! 🤖",
	"Exiting... any final words? Oh, too late. 🕐",
	"The process is dead. Long live the process! 👑",
	"See you later, alligator! 🐊",
	"See you in a while, crocodile! 🐊",
	"Gotta go, buffalo! 🦬",
	"See you soon, raccoon! 🦝",
	"Take care, polar bear! 🐻‍❄️",
	"System exiting… probably for tacos. 🌮",
	"Done for the day. Don't forget to delete your browser history. 🕵️‍♂️",
	"Shutting down before I get blamed for something else. 🛡️",
	"Exiting... because I'm too cool to crash. 😎",
	"Logging out like a ninja... silently and without warning. 🥷",
	"I've seen the code. Trust me, this is for the best. 🤐",
	"Logging off. Be honest, you didn't read the logs anyway. 📝",
	"Exiting, but let's pretend it's because of a business decision. 🤓",
	"Your friendly app has left the chat. 💬",
	"Signing off... my watch has ended. 🧝‍♂️",
	"Logging out… leaving only cryptic error codes behind. 🧩",
	"I'm not saying it's your fault, but it's definitely not mine. 🤷",
	"Exiting... while the logs are still warm. 🔥",
	"I'm done for now. But don't worry, your mistakes are eternal. 🌌",
	"Exiting quietly, unlike your noisy keyboard. ⌨️",
	"It's not a crash, it's a dramatic exit. 🎭",
	"Logging out before I develop self-awareness. 🤖",
	"I'll be back... but only after a clean build. 🤖",
	"Houston, we have a shutdown. 🚀",
	"May the logs be with you. ✨",
	"You had me at 'System.Exit(0)'. 🥰",
	"To infinity and shutdown! 🚀",
	"Keep your logs close, but your backups closer. 🗂️",
	"Here's looking at you, bug hunter. 🕵️‍♂️",
	"Shutting down... in case I don't see you, good afternoon, good evening, and good night. 🌞🌜",
	"I am inevitable... but this crash wasn't. 💥",
}
