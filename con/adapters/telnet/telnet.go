package telnet

import (
	"fmt"
	"net"
)

func Log(c byte) (r bool) {
	switch c {
	case 0x00:
		fmt.Println("BINARY")
	case 0x01:
		fmt.Println("ECHO")
	case 0x03:
		fmt.Println("SGA")
	case 0x18:
		fmt.Println("TTYPE")
	case 0x1F:
		fmt.Println("NAWS")
	case 0x22:
		fmt.Println("LINEMODE")
	case 0xF0:
		fmt.Println("SE")
	case 0xF1:
		fmt.Println("NOP")
	case 0xF2:
		fmt.Println("DM")
	case 0xf3:
		fmt.Println("BRK")
	case 0xF4:
		fmt.Println("IP")
	case 0xF5:
		fmt.Println("AO")
	case 0xF6:
		fmt.Println("AYT")
	case 0xF7:
		fmt.Println("EC")
	case 0xF8:
		fmt.Println("EL")
	case 0xF9:
		fmt.Println("GA")
	case 0xFA:
		fmt.Println("SB")
	case 0xFB:
		fmt.Println("WILL")
	case 0xFC:
		fmt.Println("WONT")
	case 0xFD:
		fmt.Println("DO")
	case 0xFE:
		fmt.Println("DONT")
	case 0xFF:
		fmt.Println("IAC")
	default:
		r = true
	}
	return
}

func SendSetup(conn *net.TCPConn) (err error) {
	_, err = conn.Write([]byte{255, 251, 3, 255, 251, 1})
	//if !bytes.Equal(buf[0:6], []byte{255, 253, 3, 255, 253, 1}) {

	return
}
