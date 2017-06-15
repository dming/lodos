package conf

import (
	"bytes"
	"os"
	"bufio"
	"strings"
	"encoding/json"
	//log "github.com/dming/lodos/mlog"
)

var (
	LenStackBuf = 1024
	LogLevel    = "debug"
	LogPath     = ""
	LogFlag     = 0
	RpcExpired  = 5 //远程访问最后期限值 单位秒[默认5秒] 这个值指定了在客户端可以等待服务端多长时间来应答
	Conf        = Config{}
)


type Rabbitmq struct {
	Uri          string
	Exchange     string
	ExchangeType string
	Queue        string
	BindingKey   string //
	ConsumerTag  string //消费者TAG
}

type Config struct {
	Modules map[string][]*ModuleSettings
	Mqtt   Mqtt
	//Master Master
}

type ModuleSettings struct {
	Id string
	Host string
	ProcessID string
	Settings  map[string]interface{}
	RabbitmqInfo *Rabbitmq
}

type Mqtt struct {
	WirteLoopChanNum int // Should > 1 	    // 最大写入包队列缓存
	ReadPackLoop     int // 最大读取包队列缓存
	ReadTimeout      int // 读取超时
	WriteTimeout     int // 写入超时
}

func LoadConfig(Path string) {
	// Read config.
	if err := readFileInto(Path); err != nil {
		panic(err)
	}

}

func readFileInto(path string) error {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//log.Info(string(data))
	return json.Unmarshal(data, &Conf)
}