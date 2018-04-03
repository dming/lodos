// Code generated by protoc-gen-go. dddDO NOT EDIT.
// source: rpc/rpc.proto

/*
Package rpcpb is a generated protocol buffer package.

It is generated from these files:
	rpc/rpc.proto

It has these top-level messages:
	RPCInfo
	ResultInfo
*/
package rpcpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package


type RPCInfo struct {
	Cid      string   `protobuf:"bytes,1,opt,name=Cid" json:"Cid,omitempty"`
	Fn       string   `protobuf:"bytes,2,opt,name=Fn" json:"Fn,omitempty"`
	ReplyTo  string   `protobuf:"bytes,3,opt,name=ReplyTo" json:"ReplyTo,omitempty"`
	Track    string   `protobuf:"bytes,4,opt,name=track" json:"track,omitempty"`
	Expired  int64    `protobuf:"varint,5,opt,name=Expired" json:"Expired,omitempty"`
	Reply    bool     `protobuf:"varint,6,opt,name=Reply" json:"Reply,omitempty"`
	ArgsType []string `protobuf:"bytes,7,rep,name=ArgsType" json:"ArgsType,omitempty"`
	Args     [][]byte `protobuf:"bytes,8,rep,name=Args,proto3" json:"Args,omitempty"`
}

func (m *RPCInfo) Reset()                    { *m = RPCInfo{} }
func (m *RPCInfo) String() string            { return proto.CompactTextString(m) }
func (*RPCInfo) ProtoMessage()               {}
func (*RPCInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *RPCInfo) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *RPCInfo) GetFn() string {
	if m != nil {
		return m.Fn
	}
	return ""
}

func (m *RPCInfo) GetReplyTo() string {
	if m != nil {
		return m.ReplyTo
	}
	return ""
}

func (m *RPCInfo) GetTrack() string {
	if m != nil {
		return m.Track
	}
	return ""
}

func (m *RPCInfo) GetExpired() int64 {
	if m != nil {
		return m.Expired
	}
	return 0
}

func (m *RPCInfo) GetReply() bool {
	if m != nil {
		return m.Reply
	}
	return false
}

func (m *RPCInfo) GetArgsType() []string {
	if m != nil {
		return m.ArgsType
	}
	return nil
}

func (m *RPCInfo) GetArgs() [][]byte {
	if m != nil {
		return m.Args
	}
	return nil
}

type ResultInfo struct {
	Cid        string `protobuf:"bytes,1,opt,name=Cid" json:"Cid,omitempty"`
	Error      string `protobuf:"bytes,2,opt,name=Error" json:"Error,omitempty"`
	ResultsType []string `protobuf:"bytes,4,rep,name=ResultType" json:"ResultType,omitempty"`
	Results     [][]byte `protobuf:"bytes,5,rep,name=Result,proto3" json:"Result,omitempty"`
}

func NewResultInfo(Cid string, Error string, ArgsType []string, Args [][]byte) *ResultInfo {
	resultInfo := &ResultInfo{
		Cid:        *proto.String(Cid),
		Error:      *proto.String(Error),
		ResultsType: ArgsType,
		Results:     Args,
	}
	return resultInfo
}
func (m *ResultInfo) Reset()                    { *m = ResultInfo{} }
func (m *ResultInfo) String() string            { return proto.CompactTextString(m) }
func (*ResultInfo) ProtoMessage()               {}
func (*ResultInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ResultInfo) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *ResultInfo) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *ResultInfo) GetResultsType() []string {
	if m != nil {
		return m.ResultsType
	}
	return nil
}

func (m *ResultInfo) GetResults() [][]byte {
	if m != nil {
		return m.Results
	}
	return nil
}

func init() {
	proto.RegisterType((*RPCInfo)(nil), "rpcpb.RPCInfo")
	proto.RegisterType((*ResultInfo)(nil), "rpcpb.ResultInfo")
}

/*
type MqCallInfo struct {
	Id	string			`protobuf:"bytes,1,opt,name=Id" json:"id,omitempty"`
	ReplyTo	string			`protobuf:"bytes,2,opt,name=ReplyTo" json:"replyTo,omitempty"`
	ArgsType []string		`protobuf:"bytes,3,opt,name=ArgsType" json:"ArgsType,omitempty"`
	Args	[][]byte		`protobuf:"bytes,4,opt,name=Args,proto3" json:"args,omitempty"`
	Flag 	string			`protobuf:"bytes,5,opt,name=Flag" json:"Flag,omitempty"`
}

func (m *MqCallInfo) Reset()                    { *m = MqCallInfo{} }
func (m *MqCallInfo) String() string            { return proto.CompactTextString(m) }
func (*MqCallInfo) ProtoMessage()               {}
func (*MqCallInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (c *MqCallInfo) GetId() string {
	if c != nil {
		return c.Id
	}
	return ""
}

func (c *MqCallInfo) GetReplyTo() string {
	if c != nil {
		return c.ReplyTo
	}
	return ""
}

func (c *MqCallInfo) GetArgsType() []string {
	if c != nil {
		return c.ArgsType
	}
	return nil
}

func (c *MqCallInfo) GetArgs() [][]byte {
	if c != nil {
		return c.Args
	}
	return nil
}

func (c *MqCallInfo) GetFlag() string {
	if c != nil {
		return c.Flag
	}
	return ""
}
*/
/*
type MqRetInfo struct {
	Flag    	string 		`protobuf:"bytes,1,opt,name=Flag" json:"Flag,omitempty"`
	Error		string 		`protobuf:"bytes,2,opt,name=Error" json:"Error,omitempty"`
	RetsType 	[]string 	`protobuf:"bytes,3,opt,name=RetType" json:"RetType,omitempty"`
	Rets    	[][]byte 	`protobuf:"bytes,4,opt,name=Rets,proto3" json:"Rets,omitempty"`
	ReplyTo 	string 		`protobuf:"bytes,5,opt,name=ReplyTo" json:"ReplyTo,omitempty"`
}

func NewMqRetInfo(flag string, err string, RetsType []string, rets [][]byte, replyTo string) *MqRetInfo {
	temp := make([]string, len(RetsType))
	for i, v := range RetsType {
		temp[i] = *proto.String(v)
	}
	ri :=  &MqRetInfo{
		Flag: 	*proto.String(flag),
		Error:	*proto.String(err),
		RetsType: temp,
		Rets:	rets,
		ReplyTo: *proto.String(replyTo),
	}
	return ri
}
func (m *MqRetInfo) Reset()                    { *m = MqRetInfo{} }
func (m *MqRetInfo) String() string            { return proto.CompactTextString(m) }
func (*MqRetInfo) ProtoMessage()               {}
func (*MqRetInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *MqRetInfo) GetFlag() string {
	if m != nil {
		return m.Flag
	}
	return ""
}

func (m *MqRetInfo) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *MqRetInfo) GetRetsType() []string {
	if m != nil {
		return m.RetsType
	}
	return nil
}

func (m *MqRetInfo) GetRets() [][]byte {
	if m != nil {
		return m.Rets
	}
	return nil
}

func (m *MqRetInfo) GetReplyTo() string {
	if m != nil {
		return m.ReplyTo
	}
	return ""
}
*/

func init() { proto.RegisterFile("rpc/rpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 230 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0xd9, 0x6c, 0xf3, 0xa7, 0x43, 0x15, 0x19, 0x8a, 0x0c, 0x1e, 0x64, 0xe9, 0x69, 0x4f,
	0x7a, 0xf0, 0x09, 0xa4, 0xb4, 0xe0, 0x4d, 0x86, 0xbe, 0x80, 0x4d, 0xa3, 0x14, 0x43, 0x76, 0x99,
	0x46, 0xb0, 0xcf, 0xe6, 0xcb, 0x49, 0x66, 0x13, 0xf1, 0xd2, 0xdb, 0xf7, 0xfb, 0x31, 0xc3, 0x7e,
	0x3b, 0x70, 0x25, 0xb1, 0x7e, 0x94, 0x58, 0x3f, 0x44, 0x09, 0x7d, 0xc0, 0x5c, 0x62, 0x1d, 0xf7,
	0xab, 0x1f, 0x03, 0x25, 0xbf, 0xae, 0x5f, 0xba, 0xf7, 0x80, 0x37, 0x60, 0xd7, 0xc7, 0x03, 0x19,
	0x67, 0xfc, 0x9c, 0x87, 0x88, 0xd7, 0x90, 0x6d, 0x3b, 0xca, 0x54, 0x64, 0xdb, 0x0e, 0x09, 0x4a,
	0x6e, 0x62, 0x7b, 0xde, 0x05, 0xb2, 0x2a, 0x27, 0xc4, 0x25, 0xe4, 0xbd, 0xbc, 0xd5, 0x9f, 0x34,
	0x53, 0x9f, 0x60, 0x98, 0xdf, 0x7c, 0xc7, 0xa3, 0x34, 0x07, 0xca, 0x9d, 0xf1, 0x96, 0x27, 0x1c,
	0xe6, 0x75, 0x95, 0x0a, 0x67, 0x7c, 0xc5, 0x09, 0xf0, 0x0e, 0xaa, 0x67, 0xf9, 0x38, 0xed, 0xce,
	0xb1, 0xa1, 0xd2, 0x59, 0x3f, 0xe7, 0x3f, 0x46, 0x84, 0xd9, 0x90, 0xa9, 0x72, 0xd6, 0x2f, 0x58,
	0xf3, 0xaa, 0x05, 0xe0, 0xe6, 0xf4, 0xd5, 0xf6, 0x17, 0xfa, 0x2f, 0x21, 0xdf, 0x88, 0x04, 0x19,
	0xbf, 0x90, 0x00, 0xef, 0xa7, 0x2d, 0x7d, 0x27, 0x15, 0xfe, 0x67, 0xf0, 0x16, 0x8a, 0x44, 0x5a,
	0x7a, 0xc1, 0x23, 0xed, 0x0b, 0xbd, 0xdc, 0xd3, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x2f, 0x95,
	0xb9, 0x9e, 0x4a, 0x01, 0x00, 0x00,
}
