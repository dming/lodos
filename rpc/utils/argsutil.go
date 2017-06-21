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
package argsutils

import (
	"fmt"
	"reflect"
	"github.com/dming/lodos/module"
	"github.com/dming/lodos/utils"
)

const (
	NULL string = 	"null"	//nil   null
	BOOL string  = 	"bool"	//bool
	INT string = 	"int"	//int
	INT32 string = 	"int32" //int32
	LONG string =	"long"	//long64
	FLOAT string =	"float"	//float32
	DOUBLE string =	"double"	//float64
	BYTES string =	"bytes"	//[]byte
	STRING string =	"string" //string
	MAP string =	"map"	//map[string]interface{}
	MAPSTR string =	"mapstr"	//map[string]string{}
)

func ArgsTypeAnd2Bytes(app module.AppInterface, arg interface{}) (string,[]byte,error) {
	switch v2:=arg.(type) {
	case nil:
		return NULL,	nil,nil
	case string:
		return STRING,	[]byte(v2),nil
	case bool:
		return BOOL,	utils.BoolToBytes(v2),nil
	case int32:
		return INT32,	utils.Int32ToBytes(v2),nil
	case int:
		return INT,		utils.IntToBytes(v2),nil
	case int64:
		return LONG,	utils.Int64ToBytes(v2),nil
	case float32:
		return FLOAT,	utils.Float32ToBytes(v2),nil
	case float64:
		return DOUBLE,	utils.Float64ToBytes(v2),nil
	case []byte:
		return BYTES,	v2,nil
	case map[string]interface{}:
		bytes, err := utils.MapToBytes(v2)
		if err != nil{
			return MAP,	nil,err
		}
		return MAP,	bytes,nil
	case map[string]string:
		bytes,err:=utils.MapToBytesString(v2)
		if err != nil{
			return MAPSTR,nil,err
		}
		return MAPSTR,	bytes,nil
	default:
		if app != nil {
			for _, v := range app.GetRPCSerialize() {
				ptype, vk, err := v.Serialize(arg)
				if err == nil {
					//解析成功了
					return ptype, vk, err
				}
			}
		}
		return NULL, nil,fmt.Errorf("ArgsTypeAnd2Bytes: args [%s] Types not allowed",reflect.TypeOf(arg))
	}
}

func Bytes2Args (app module.AppInterface, argsType string,args []byte ) (interface{}, error){
	switch argsType {
	case NULL:
		return nil,nil
	case STRING:
		return string(args),nil
	case BOOL:
		return utils.BytesToBool(args),nil
	case INT:
		return utils.BytesToInt(args),nil
	case INT32:
		return utils.BytesToInt32(args),nil
	case LONG:
		return utils.BytesToInt64(args),nil
	case FLOAT:
		return utils.BytesToFloat32(args),nil
	case DOUBLE:
		return utils.BytesToFloat64(args),nil
	case BYTES:
		return args,nil
	case MAP:
		mps,errs:= utils.BytesToMap(args)
		if errs!=nil{
			return	nil,errs
		}
		return mps,nil
	case MAPSTR:
		mps,errs:= utils.BytesToMapString(args)
		if errs!=nil{
			return	nil,errs
		}
		return mps,nil
	default:
		if app != nil {
			for _, v := range app.GetRPCSerialize() {
				vk, err := v.Deserialize(argsType, args)
				if err == nil {
					//解析成功了
					return vk, err
				}
			}
		}
		return	nil,fmt.Errorf("Bytes2Args: args [%s] Types not allowed",argsType)
	}
}

