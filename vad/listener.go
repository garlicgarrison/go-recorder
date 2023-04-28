package vad

type VADListener interface {
	OnSpeechDetected()
	OnSilenceDetected()
}
