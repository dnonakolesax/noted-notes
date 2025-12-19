package ws

const blockIDLen = 36 // длина UUID-строки

// разделить входящий бинарный кадр
func splitFrame(msg []byte) (blockID string, payload []byte, ok bool) {
    if len(msg) < blockIDLen {
        return "", nil, false
    }
    return string(msg[:blockIDLen]), msg[blockIDLen:], true
}

// собрать кадр для отправки
func makeFrame(blockID string, payload []byte) []byte {
    idBytes := []byte(blockID) // предполагаем, что len == 36
    buf := make([]byte, len(idBytes)+len(payload))
    copy(buf, idBytes)
    copy(buf[len(idBytes):], payload)
    return buf
}
