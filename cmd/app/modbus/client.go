package modbus

import (
	"fmt"
	"net"
	"strings"

	modbus "github.com/things-go/go-modbus"
	tags "zheleznovux.com/modbus-console/cmd/storage"
)

type Client struct {
	name    string
	ip      string
	port    int
	slaveId uint8
	tags    []tags.BaseTag
	sender  modbus.Client
}

func (c *Client) Name() string {
	return c.name
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

func NewClinet(ip string, port int, slaveID uint8, name string) (Client, error) {
	var c Client

	c.name = name
	c.slaveId = slaveID

	ipAddr := net.ParseIP(strings.TrimSpace(ip))
	if ipAddr == nil {
		return Client{}, fmt.Errorf("Ошибка парса Ip {constructor Clinet}")
	} else {
		c.ip = ip
	}

	if (port > 65536) || (port < 1) {
		c.port = 502
		return Client{}, fmt.Errorf("Ошибка парса порта {constructor Clinet}")
	} else {
		c.port = port
	}
	c.tags = make([]tags.BaseTag, 0)

	// if err := c.connect(); err != nil {
	// 	return Client{}, err
	// }

	return c, nil
}

func (c *Client) SetTag(name string, address string, scanPeriod float64, typeTag string, bit uint8) error {
	if typeTag == "coil" {
		var tag tags.CoilTag
		tag.SetName(name)
		tag.SetDataType(typeTag)

		err := tag.SetAddress(address)
		if err != nil {
			return fmt.Errorf("SetAddress")
		}

		err = tag.SetScanPeriod(scanPeriod)
		if err != nil {
			return fmt.Errorf("SetScanPeriod")
		}

		err = tag.SetBit(bit)
		if err != nil {
			return fmt.Errorf("SetBit")
		}
		c.tags = append(c.tags, &tag)
	} else {
		var tag tags.WordTag
		tag.SetName(name)
		tag.SetDataType(typeTag)

		err := tag.SetAddress(address)
		if err != nil {
			return fmt.Errorf("SetAddress")
		}

		err = tag.SetScanPeriod(scanPeriod)
		if err != nil {
			return fmt.Errorf("SetScanPeriod")
		}
		c.tags = append(c.tags, &tag)
	}
	return nil
}

func (c *Client) Tags() []tags.BaseTag {
	return c.tags
}

func (c *Client) SetIp(ip string) error {
	ipAddr := net.ParseIP(strings.TrimSpace(ip))
	if ipAddr == nil {
		return fmt.Errorf("Недопустимый Ip. {setter Ip}")
	} else {
		c.ip = ip
		return nil
	}
}

func (c *Client) SetPort(port int) error {
	if (port > 65536) || (port < 1) {
		c.port = 502
		return fmt.Errorf("Ошибка парса порта {setter Port}")
	} else {
		c.port = port
		return nil
	}
}

func (c *Client) Connect() error {
	// provider := modbus.NewTCPClientProvider(
	// 	c.ip+":"+fmt.Sprint(c.port),
	// 	modbus.WithEnableLogger())
	provider := modbus.NewTCPClientProvider(
		c.ip + ":" + fmt.Sprint(c.port))
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
		return fmt.Errorf("sender nil")
	}

	resp, err := c.sender.ReadHoldingRegisters(c.slaveId, c.tags[id].GetAddress(), 1)
	if err != nil {
		c.tags[id].SetState(false)
		return err
	}

	if len(resp) > 0 {
		c.tags[id].ReadSending(resp[0])
		return nil
	}

	return fmt.Errorf("resp nil")

}
