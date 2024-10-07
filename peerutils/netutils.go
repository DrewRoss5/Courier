package peerutils

import "net"

const BUF_SIZE = 1024

// recieves data of unknown size from Conn object
func RecvAll(conn net.Conn) (int, []byte, error) {
	message := make([]byte, 0, BUF_SIZE)
	buf := make([]byte, BUF_SIZE)
	totalRead := 0
	for {
		sizeRead, err := conn.Read(buf)
		if sizeRead == 0 {
			break
		}
		if err != nil {
			return 0, nil, err
		}
		totalRead += sizeRead
		message = append(message, buf[:sizeRead]...)
	}
	return totalRead, message, nil
}
