package conf

import (
	"bytes"
	"os"
	"bufio"
	"strings"
	"encoding/json"

	"fmt"
	"io/ioutil"
)

var (
	LenStackBuf = 1024
	LogLevel    = "debug"
	LogPath     = ""
	LogFlag     = 0
	RpcExpired  = 5 //远程访问最后期限值 单位秒[默认5秒] 这个值指定了在客户端可以等待服务端多长时间来应答
	Conf        = Config{}
)

type Config struct {
	Log map[string]interface{}
	RPC RPC
	Modules map[string][]*ModuleSettings
	Mqtt   Mqtt
	//Master Master
	Settings map[string]interface{}
}


type Rabbitmq struct {
	Uri          string
	Exchange     string
	ExchangeType string
	Queue        string
	BindingKey   string //
	ConsumerTag  string //消费者TAG
}

/*
type RedisInfo struct {
	redis.Options
}*/

type Redis struct {
	Uri string  //redis://:[password]@[ip]:[port]/[db]
	Queue string
}

type RPC struct {
	MaxCoroutine int //模块同时可以创建的最大协程数量默认是100
	RpcExpired int	//远程访问最后期限值 单位秒[默认5秒] 这个值指定了在客户端可以等待服务端多长时间来应答
	LogSuccess bool	//是否打印请求处理成功的日志
	Log bool		//是否打印RPC的日志
}

type ModuleSettings struct {
	Id string
	Host string
	ProcessID string
	Settings  map[string]interface{}
	RabbitmqInfo *Rabbitmq
	//redisInfo *RedisInfo
	RedisInfo *Redis
}

type Mqtt struct {
	WirteLoopChanNum int // Should > 1 	    // 最大写入包队列缓存
	ReadPackLoop     int // 最大读取包队列缓存
	ReadTimeout      int // 读取超时
	WriteTimeout     int // 写入超时
}

type SSH struct {
	Host     string
	Port     int
	User     string
	Password string
}

/**
host:port
*/
func (s *SSH) GetSSHHost() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Process struct {
	ProcessID string
	Host      string
	//执行文件
	Execfile string
	//日志文件目录
	//pid.nohup.log
	//pid.access.log
	//pid.error.log
	LogDir string
	//自定义的参数
	Args map[string]interface{}
}

type Master struct {
	Enable  bool
	WebRoot string
	WebHost string
	SSH     []*SSH
	Process []*Process
}

func (m *Master) GetSSH(host string) *SSH {
	for _, ssh := range m.SSH {
		if ssh.Host == host {
			return ssh
		}
	}
	return nil
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

// If read the file has an error,it will throws a panic.
func fileToStruct(path string, ptr *[]byte) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	*ptr = data
}
