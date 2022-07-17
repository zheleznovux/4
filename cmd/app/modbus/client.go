package modbus

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	modbus "github.com/things-go/go-modbus"
	constants "zheleznovux.com/modbus-console/cmd/constants"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

type Client struct {
	name               string
	ip                 string
	port               int
	slaveId            uint8
	isDebug            bool
	connectionAttempts int
	connectionTimeout  time.Duration
	tags               []tags.TagInterface
	sender             modbus.Client
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) ConnectionAttempts() int {
	return c.connectionAttempts
}

func (c *Client) Ip() string {
	return c.ip
}

func (c *Client) Port() int {
	return c.port
}

func (c *Client) SalveId() uint8 {
	return c.slaveId
}

func (c *Client) ConnectionTimeout() time.Duration {
	return c.connectionTimeout
}

func NewClinet(ip string, port int, slaveID uint8, name string, debug bool, ConnectionAttempts int, ConnectionTimeout float64) (Client, error) {
	var c Client

	c.name = name
	c.slaveId = slaveID
	c.isDebug = debug
	c.connectionTimeout = time.Duration(ConnectionTimeout * float64(time.Second))
	c.connectionAttempts = ConnectionAttempts

	ipAddr := net.ParseIP(strings.TrimSpace(ip))
	if ipAddr == nil {
		return Client{}, fmt.Errorf("ошибка парса Ip {constructor сlinet}")
	} else {
		c.ip = ip
	}

	if (port > 65536) || (port < 1) {
		c.port = 502
		return Client{}, fmt.Errorf("ошибка парса порта {constructor сlinet}")
	} else {
		c.port = port
	}
	c.tags = make([]tags.TagInterface, 0)

	return c, nil
}

func (c *Client) SetTag(name string, address uint16, scanPeriod float64, typeTag string) error {
	switch typeTag {
	case constants.COIL_TYPE:
		{
			var tag tags.CoilTag
			tag.SetName(name)
			tag.SetDataType()

			tag.SetAddress(address)

			err := tag.SetScanPeriod(scanPeriod)
			if err != nil {
				return fmt.Errorf("SetScanPeriod")
			}

			c.tags = append(c.tags, &tag)
		}
	case constants.WORD_TYPE:
		{
			var tag tags.WordTag
			tag.SetName(name)
			tag.SetDataType()

			tag.SetAddress(address)

			err := tag.SetScanPeriod(scanPeriod)
			if err != nil {
				return fmt.Errorf("SetScanPeriod")
			}
			c.tags = append(c.tags, &tag)

		}
	case constants.DWORD_TYPE:
		{
			var tag tags.DWordTag
			tag.SetName(name)
			tag.SetDataType()

			tag.SetAddress(address)

			err := tag.SetScanPeriod(scanPeriod)
			if err != nil {
				return fmt.Errorf("SetScanPeriod")
			}
			c.tags = append(c.tags, &tag)

		}
	default:
		return fmt.Errorf("SetTag parsErr datatype")
	}
	return nil
}

func (c *Client) Tags() []tags.TagInterface {
	return c.tags
}

func (c *Client) SetIp(ip string) error {
	ipAddr := net.ParseIP(strings.TrimSpace(ip))
	if ipAddr == nil {
		return fmt.Errorf("недопустимый Ip. {setter Ip}")
	} else {
		c.ip = ip
		return nil
	}
}

func (c *Client) SetPort(port int) error {
	if (port > 65536) || (port < 1) {
		c.port = 502
		return fmt.Errorf("ошибка парса порта {setter Port}")
	} else {
		c.port = port
		return nil
	}
}

func (c *Client) Connect() error {
	var provider modbus.ClientProvider

	if c.isDebug {
		provider = modbus.NewTCPClientProvider(
			c.ip+":"+fmt.Sprint(c.port),
			modbus.WithEnableLogger())
	} else {
		provider = modbus.NewTCPClientProvider(
			c.ip + ":" + fmt.Sprint(c.port))
	}
	c.sender = modbus.NewClient(provider)

	//устанавливаем соединение
	err := c.sender.Connect()
	if err != nil {
		return fmt.Errorf("Ошибка соединения! " + err.Error())
	} else {
		return nil
	}
}

func (c *Client) Send(id int) error {
	if c.sender == nil {
		c.tags[id].SetState(false)
		return errors.New("sender nil")
	}
	//новые типы должны быть указаны здесь
	switch c.tags[id].DataType() {
	case constants.COIL_TYPE:
		{
			resp, err := c.sender.ReadDiscreteInputs(c.slaveId, c.tags[id].Address(), 1)

			if err != nil {
				c.tags[id].SetState(false)
				return err
			}

			if len(resp) > 0 {
				c.tags[id].(*tags.CoilTag).SetValue(resp[0])
				return nil
			} else {
				return errors.New("response nil")
			}
		}
	case constants.WORD_TYPE:
		{
			resp, err := c.sender.ReadHoldingRegisters(c.slaveId, c.tags[id].Address(), 1)

			if err != nil {
				c.tags[id].SetState(false)
				return err
			}

			if len(resp) > 0 {
				c.tags[id].(*tags.WordTag).SetValue(resp[0])
				return nil
			} else {
				return errors.New("response nil")
			}
		}
	case constants.DWORD_TYPE:
		{
			resp, err := c.sender.ReadHoldingRegisters(c.slaveId, c.tags[id].Address(), 2)

			if err != nil {
				c.tags[id].SetState(false)
				return err
			}

			if len(resp) > 1 {
				var tmp uint32 = (uint32(resp[0]) << 16) + uint32(resp[1])
				c.tags[id].(*tags.DWordTag).SetValue(tmp)
				return nil
			} else {
				return errors.New("response nil")
			}
		}
	default:
		return errors.New("resp nil")
	}
}
