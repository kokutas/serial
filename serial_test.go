package serial

import (
	"log"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Ltime)
}

func TestNew(t *testing.T) {
	sp, err := New("COM3", 9600, 8, 1, PARITY_NONE, time.Duration(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	if err := sp.Open(); err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 500; i++ {
		if err := sp.Flush(); err != nil {
			log.Fatal(err)
		}
		data := []byte{0x01, 0x03, 0x00, 0x01, 0x00, 0x01, 0xD5, 0xCA}
		_, err = sp.Write(data)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(40 * time.Millisecond)
		buf := make([]byte, 128)
		n, err := sp.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		raw := buf[:n]
		log.Printf("index = %d, %x\r\n", i, raw)
		if n > 3 {
			log.Printf("index = %d, buf_len = %d , raw_data = 0x%x , data_len = %x , data = %x", i, n, raw, raw[2:3], raw[3:n-2])
		}
	}
}
