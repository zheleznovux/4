package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	modbus1 "github.com/things-go/go-modbus"
	constants "zheleznovux.com/modbus-console/cmd/serverStorage/constants"
	"zheleznovux.com/modbus-console/cmd/serverStorage/logger"
	tag "zheleznovux.com/modbus-console/cmd/serverStorage/tag"
)

type ClientModbus struct {
	name               string
	connectionType     string
	ip                 string
	port               int
	slaveId            uint8
	connectionAttempts int
	connectionTimeout  time.Duration
	log                logger.Logger
	state              string
	tags               []tag.TagInterface
	sender             modbus1.Client
}

func (c *ClientModbus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name    string
		Ip      string
		SlaveId uint8
		State   string
		Tags    []tag.TagInterface
	}{
		Name:    c.name,
		Ip:      c.ip,
		SlaveId: c.slaveId,
		State:   c.state,
		Tags:    c.tags,
	})
}

// ======================инициализация========================{

// =================================Name
func (c *ClientModbus) Name() string {
	return c.name
}
func (c *ClientModbus) SetName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("invalid client name. {setter client name}")
	}
	c.name = name
	return nil
}

// =================================Type
func (c *ClientModbus) Type() string {
	return c.connectionType
}
func (c *ClientModbus) SetType() {
	c.connectionType = constants.MODBUS_TCP
}

// ===================================IP
func (c *ClientModbus) Ip() string {
	return c.ip
}

// using net.parseIp
func (c *ClientModbus) SetIp(ip string) error {
	ipAddr := net.ParseIP(strings.TrimSpace(ip))
	if ipAddr == nil {
		return errors.New("invalid client Ip. {setter client Ip}")
	} else {
		c.ip = ip
		return nil
	}
}

// ===================ConnectionAttempts
func (c *ClientModbus) ConnectionAttempts() int {
	return c.connectionAttempts
}
func (c *ClientModbus) SetConnectionAttempts(ca int) error {
	if ca <= 0 {
		return errors.New("invalid client connection attempts. {setter client connection attempts}")
	}
	c.connectionAttempts = ca
	return nil
}

// =================================Port
func (c *ClientModbus) Port() int {
	return c.port
}
func (c *ClientModbus) SetPort(port int) error {
	if (port > 0xFFFF) || (port < 0) {
		c.port = 502
		return errors.New("invalid client port. {setter client port}")
	} else {
		c.port = port
		return nil
	}
}

// ==============================SlaveID
func (c *ClientModbus) SalveId() uint8 {
	return c.slaveId
}
func (c *ClientModbus) SetSalveId(sid uint8) error {
	if sid > 0xFF {
		return errors.New("invalid client slaveID. {setter client slaveID}")
	}
	c.slaveId = sid
	return nil
}

// =================================Tags
func (c *ClientModbus) Tags() []tag.TagInterface {
	return c.tags
}
func (c *ClientModbus) SetTags(tags []tag.TagInterface) error {
	for id := range tags {
		if _, err := c.TagByName(tags[id].Name()); err != nil {
			return err
		}
	}
	c.tags = tags
	return nil
}

// ============================TagById
func (c *ClientModbus) TagById(id int) (tag.TagInterface, error) {
	if (id >= len(c.tags)) || (id < 0) {
		return nil, errors.New("invalid id client tag. {getter client tag by id}")
	}
	return c.tags[id], nil
}
func (c *ClientModbus) TagByName(name string) (tag.TagInterface, error) {
	for id := range c.tags {
		if c.tags[id].Name() == name {
			return c.tags[id], nil
		}
	}
	return nil, errors.New("invalid client tag name. {getter client tag by name}")
}
func (c *ClientModbus) SetTag(t tag.TagInterface) error {
	if _, err := c.TagByName(t.Name()); err != nil {
		c.tags = append(c.tags, t)
		return nil
	}
	return errors.New("client tag name already exists. {setter client tag}")
}

// ====================ConnectionTimeout
func (c *ClientModbus) ConnectionTimeout() time.Duration {
	return c.connectionTimeout
}
func (c *ClientModbus) SetConnectionTimeout(s float64) error {
	if s < 0 {
		return errors.New("client connection timeout < 0. {setter connection timeout}")
	}
	c.connectionTimeout = time.Duration(s) * time.Second
	return nil
}

// // ==============================State
func (c *ClientModbus) setState(state string) (isChanged bool) {
	if c.state != state {
		c.state = state
		isChanged = true
	} else {
		isChanged = false
	}
	return
}

// ==========================Constructor
func NewClinetModbus(ip string, port int, slaveID uint8, name string, debug bool, ConnectionAttempts int, ConnectionTimeout float64) (*ClientModbus, error) {
	var c ClientModbus

	err := c.SetName(name)
	if err != nil {
		return nil, err
	}
	c.SetType()
	err = c.SetIp(ip)
	if err != nil {
		return nil, err
	}
	err = c.SetPort(port)
	if err != nil {
		return nil, err
	}
	err = c.SetSalveId(slaveID)
	if err != nil {
		return nil, err
	}
	err = c.SetConnectionAttempts(ConnectionAttempts)
	if err != nil {
		return nil, err
	}
	err = c.SetConnectionTimeout(ConnectionTimeout)
	if err != nil {
		return nil, err
	}
	c.log = logger.Logger{
		ParentNodeName: c.name,
		ParentNodeIp:   c.ip,
		IsLogOutput:    debug,
	}

	return &c, nil
}

//======================инициализация========================}

func (c *ClientModbus) Connect() error {
	provider := modbus1.NewTCPClientProvider(
		c.ip + ":" + fmt.Sprint(c.port))
	c.sender = modbus1.NewClient(provider)
	defer c.sender.Close()

	//устанавливаем соединение
	err := c.sender.Connect()
	if err != nil {
		return errors.New("Ошибка соединения! " + err.Error())
	} else {
		return nil
	}
}

func (c *ClientModbus) Send(id int) error {
	if c.sender == nil {
		c.tags[id].SetState(false)
		return errors.New("sender nil")
	}
	defer c.sender.Close()

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
				c.tags[id].(*tag.CoilTag).SetValue(resp[0])
				// c.log.WriteWithTag(logger.INFO, c.state, c.tags[id].Name(), "Значение: "+strconv.Itoa(int(resp[0]))+".")
				return nil
			} else {
				c.tags[id].SetState(false)
				// c.log.WriteWithTag(logger.WARNING, c.state, c.tags[id].Name(), "Значение не было считано!")
				return nil
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
				c.tags[id].(*tag.WordTag).SetValue(resp[0])
				// c.log.WriteWithTag(logger.INFO, c.state, c.tags[id].Name(), "Значение: "+strconv.Itoa(int(resp[0]))+".")
				return nil
			} else {
				c.tags[id].SetState(false)
				// c.log.WriteWithTag(logger.WARNING, c.state, c.tags[id].Name(), "Значение не было считано!")
				return nil
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
				c.tags[id].(*tag.DWordTag).SetValue(tmp)
				// c.log.WriteWithTag(logger.INFO, c.state, c.tags[id].Name(), "Значение: "+strconv.Itoa(int(tmp))+".")
				return nil
			} else {
				c.tags[id].SetState(false)
				// c.log.WriteWithTag(logger.WARNING, c.state, c.tags[id].Name(), "Значение не было считано!")
				return nil
			}
		}
	default:
		return errors.New("resp nil")
	}
}

func (c *ClientModbus) Start(stop chan struct{}, wg *sync.WaitGroup) {
	connection := make(chan bool)
	quit := make(chan struct{})
	defer wg.Done()

	var wgi sync.WaitGroup
	wgi.Add(1)
	go c.TryConnect(stop, connection, &wgi)
	for {
		select {
		case <-stop: //канал сверху. Завершение сессии
			{
				close(quit)
				wgi.Wait()
				return
			}
		case cb := <-connection: // канал снизу. Плохое подключение => реконект
			{
				if cb {
					close(quit)
					wgi.Wait()
					quit = make(chan struct{})
					for tagId := range c.tags {
						wgi.Add(1)
						go c.startSender(tagId, quit, &wgi, connection)
					}
				} else {
					close(quit)
					wgi.Wait()
					quit = make(chan struct{})
					wgi.Add(1)
					go c.TryConnect(stop, connection, &wgi)
				}
			}
		}
	}

}

func (c *ClientModbus) TryConnect(quit chan struct{}, connection chan bool, wg *sync.WaitGroup) { /// connection day out
	defer wg.Done()
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-quit:
			{
				return
			}
		case <-ticker.C:
			{
				ticker.Stop()
				for i := 1; i <= c.connectionAttempts; i++ {
					select {
					case <-quit:
						{
							return
						}
					default:
						{
							err := c.Connect()
							if err == nil {
								if isChanged := c.setState(constants.GOOD); isChanged {
									c.log.Write(logger.INFO, c.state, "Подключенно!")
								}
								connection <- true
								return
							}
						}
					}
				}
				if isChanged := c.setState(constants.BAD); isChanged {
					c.log.Write(logger.WARNING, c.state, "Неудалось подключиться!")
				}

				ticker.Reset(c.connectionTimeout)
			}
		}
	}
}

func (c *ClientModbus) startSender(tagId int, quit chan struct{}, wg *sync.WaitGroup, connect chan bool) {
	// c.log.WriteWithTag(logger.INFO, c.state, c.tags[tagId].Name(), "Запущен опрос тега!")

	ticker := time.NewTicker(time.Duration(c.tags[tagId].ScanPeriod()) * time.Second)

	defer wg.Done()

	for {
		select {
		case <-quit:
			{
				// c.log.WriteWithTag(logger.INFO, c.state, c.tags[tagId].Name(), "Завершен опрос тега!")
				return
			}
		case <-ticker.C:
			{
				err := c.Send(tagId)
				if err != nil {
					c.log.WriteWithTag(logger.ERROR, c.state, c.tags[tagId].Name(), err.Error())
					connect <- false
				}
			}
		}
	}
}
