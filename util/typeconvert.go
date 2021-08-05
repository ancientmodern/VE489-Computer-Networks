package util

func Bool2Byte(val bool) byte {
	if val {
		return 1
	} else {
		return 0
	}
}

func Byte2Bool(val byte) bool {
	if val == 0 {
		return false
	} else {
		return true
	}
}

func Bool2Int(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}

func Int2Bool(val int) bool {
	if val == 0 {
		return false
	} else {
		return true
	}
}
