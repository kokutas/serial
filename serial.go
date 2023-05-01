package serial

import (
	"errors"
	"time"
)

// 校验位类型
type Parity byte

// 校验位常量
const (
	PARITY_NONE  Parity = 'N' // 无校验(NONE): 没有校验位
	PARITY_ODD   Parity = 'O' // 奇偶校验位-奇校验(ODD): 让传输的数据(包含校验位)中1的个数为奇数;如果传输字节中1的个数是偶数,则校验位为"1",奇数相反.
	PARITY_EVEN  Parity = 'E' // 奇偶校验位-偶校验(EVEN): 让传输的数据(包含校验位)中1的个数为偶数;如果传输字节中1的个数是偶数,则校验位为"0",奇数相反.
	PARITY_MARK  Parity = 'M' // 固定校验位(Stick)-1校验(MARK): 校验位总为1
	PARITY_SPACE Parity = 'S' // 固定校验位(Stick)-0校验(SPACE): 校验位总为0
)

// 停止位类型
type StopBits byte

// 停止位常量
const (
	STOP_ONE      StopBits = 1
	STOP_ONE_HALF StopBits = 15 // 1.5
	STOP_TWO      StopBits = 2
)

// 默认值
const (
	DEFAULT_BAUD_RATE uint     = 9600        // 默认波特率
	DEFAULT_PARITY    Parity   = PARITY_NONE // 默认校验位
	DEFAULT_DATA_BITS uint8    = 8           // 默认数据位
	DEFAULT_STOP_BITS StopBits = STOP_ONE    // 默认停止位
)

var (
	// ErrBadDataBits is returned if DataBits is not supported.
	ErrBadDataBits error = errors.New("unsupported serial data bits")
	// ErrBadStopBits is returned if the specified StopBits setting not supported.
	ErrBadStopBits error = errors.New("unsupported stop bit setting")
	// ErrBadParity is returned if the parity is not supported.
	ErrBadParity error = errors.New("unsupported parity setting")
)

// 串口配置
type serial struct {
	Address  string        `json:"address" validate:"required"`                     // 串口地址,例如:/dev/ttyS0
	BaudRate uint          `json:"baud_rate" validate:"required"`                   // 波特率: 默认9600
	DataBits uint8         `json:"data_bits" validate:"required,oneof=5 6 7 8"`     // 数据位: 5,6,7,8,默认8
	StopBits StopBits      `json:"stop_bits" validate:"required,oneof=1 2"`         // 停止位: 1,2.默认1
	Parity   Parity        `json:"parity" validate:"required,oneof=78 79 69 77 83"` // 校验位: N-None(78),O-Odd(79),E-Even(69),M-Mark(77),S-Space(83)
	Timeout  time.Duration `json:"timeout" validate:"-"`                            // 读/写超时时间
	port     *port         `json:"-" validate:"-"`
	// RTSFlowControl bool
	// DTRFlowControl bool
	// XONFlowControl bool
	// CRLFTranslate bool
}

func New(address string, baudRate uint, dataBits uint8, stopBits StopBits, parity Parity, timeout time.Duration) (*serial, error) {
	serial := &serial{
		Address:  address,
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: stopBits,
		Parity:   parity,
		Timeout:  timeout,
	}
	if serial.BaudRate == 0 {
		serial.BaudRate = DEFAULT_BAUD_RATE
	}
	if serial.Parity == 0 {
		serial.Parity = DEFAULT_PARITY
	}
	if serial.DataBits == 0 {
		serial.DataBits = DEFAULT_DATA_BITS
	}
	if serial.StopBits == 0 {
		serial.StopBits = DEFAULT_STOP_BITS
	}
	if err := Validate(serial, "en", true); err != nil {
		return nil, err
	}
	return serial, nil
}
func (s *serial) Open() (err error) {
	s.port, err = open(s.Address, s.BaudRate, s.DataBits, s.StopBits, s.Parity, s.Timeout)
	return
}
func (s *serial) Close() error {
	if s.port != nil {
		return s.port.close()
	}
	return nil
}
func (s *serial) Flush() error {
	if s.port != nil {
		return s.port.flush()
	}
	return nil
}

func (s *serial) Write(buf []byte) (int, error) {
	if s.port != nil {
		return s.port.write(buf)
	}
	return 0, errors.New("serial port is not allowed empty")
}

func (s *serial) Read(buf []byte) (int, error) {
	if s.port != nil {
		return s.port.read(buf)
	}
	return 0, errors.New("serial port is not allowed empty")
}
