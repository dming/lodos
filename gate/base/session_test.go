// Copyright 2014 loolgame Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package basegate
import (
	"github.com/golang/protobuf/proto"
	"testing"
)
func TestSession(t *testing.T) {
	s := &Sessionpb{        // 使用辅助函数设置域的值
		IP: *proto.String("127.0.0.1"),
		Network:  *proto.String("tcp"),
		Sessionid:  *proto.String("iii"),
		Serverid:  *proto.String("232244"),
	}    // 进行编码
	s.Settings=map[string]string{"isLogin":"true"}
	data, err := proto.Marshal(s)
	if err != nil {
		t.Fatalf("marshaling error: ", err)
	}    // 进行解码
	newSessionpb := &Sessionpb{}
	err = proto.Unmarshal(data, newSessionpb)
	if err != nil {
		t.Fatalf("unmarshaling error: ", err)
	}    // 测试结果
	if s.Serverid != newSessionpb.GetServerid() {
		t.Fatalf("data mismatch %q != %q", s.GetServerid(), newSessionpb.GetServerid())
	}
	if newSessionpb.GetSettings()==nil{
		t.Fatalf("data mismatch Settings == nil")
	}else{
		if newSessionpb.GetSettings()["isLogin"]!="true"{
			t.Fatalf("data mismatch %q != %q", s.GetSettings()["isLogin"], newSessionpb.GetSettings()["isLogin"])
		}
	}

}
