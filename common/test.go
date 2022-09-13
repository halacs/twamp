package common

/*
Function header called when a test package arrived back.
Can be used to show some progress
*/
type TwampTestCallbackFunction func(targetPackets int, result *TwampResult, stats *PingResultStats)
