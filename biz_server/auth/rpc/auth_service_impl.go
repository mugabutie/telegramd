/*
 *  Copyright (c) 2017, https://github.com/nebulaim
 *  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpc

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"../../../biz_model/dal/dao"
	"../../../biz_model/dal/dataobject"
	"../../../frontend/id"
	"../../../grpc_util"
	"../../../mtproto"
	"github.com/ttacon/libphonenumber"
	"golang.org/x/net/context"
	"time"
)

type AuthServiceImpl struct {
}

func (s *AuthServiceImpl) AuthLogOut(ctx context.Context, request *mtproto.TLAuthLogOut) (*mtproto.Bool, error) {
	glog.Infof("AuthLogOut - Process: {%v}", request)

	// TODO(@benqi): Logout逻辑处理，失效AuthKey
	reply := mtproto.MakeBool(&mtproto.TLBoolTrue{})

	glog.Infof("AuthLogOut - reply: {%v}\n", reply)
	return reply, nil
}

func (s *AuthServiceImpl) AuthResetAuthorizations(ctx context.Context, request *mtproto.TLAuthResetAuthorizations) (*mtproto.Bool, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthSendInvites(ctx context.Context, request *mtproto.TLAuthSendInvites) (*mtproto.Bool, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthBindTempAuthKey(ctx context.Context, request *mtproto.TLAuthBindTempAuthKey) (*mtproto.Bool, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthCancelCode(ctx context.Context, request *mtproto.TLAuthCancelCode) (*mtproto.Bool, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthDropTempAuthKeys(ctx context.Context, request *mtproto.TLAuthDropTempAuthKeys) (*mtproto.Bool, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

// 检查手机号码是否已经注册
func (s *AuthServiceImpl) AuthCheckPhone(ctx context.Context, request *mtproto.TLAuthCheckPhone) (*mtproto.Auth_CheckedPhone, error) {
	glog.Infof("AuthCheckPhone - Process: {%v}", request)

	// TODO(@benqi): panic/recovery
	usersDAO := dao.GetUsersDAO(dao.DB_SLAVE)

	// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumer := libphonenumber.NormalizeDigitsOnly(request.PhoneNumber)

	usersDO := usersDAO.SelectByPhoneNumber(phoneNumer)

	var reply *mtproto.Auth_CheckedPhone
	if usersDO == nil {
		// 未注册
		reply = mtproto.MakeAuth_CheckedPhone(&mtproto.TLAuthCheckedPhone{
			PhoneRegistered: mtproto.MakeBool(&mtproto.TLBoolFalse{}),
		})
	} else {
		// 已经注册
		reply = mtproto.MakeAuth_CheckedPhone(&mtproto.TLAuthCheckedPhone{
			PhoneRegistered: mtproto.MakeBool(&mtproto.TLBoolTrue{}),
		})
	}

	glog.Infof("AuthCheckPhone - reply: {%v}\n", reply)

	// TODO(@benqi): 强制验证通过
	// checked := &mtproto.TLAuthCheckedPhone{mtproto.ToBool(true)}
	return reply, nil
}

func (s *AuthServiceImpl) AuthSendCode(ctx context.Context, request *mtproto.TLAuthSendCode) (*mtproto.Auth_SentCode, error) {
	glog.Infof("AuthSendCode - Process: {%v}", request)

	// Check TLAuthSendCode
	// CurrentNumber: 是否为本机电话号码
	// 检查数据是否合法
	//switch request.CurrentNumber.(type) {
	//case *mtproto.Bool_BoolFalse:
	//	// 本机电话号码，AllowFlashcall为false
	//	if request.AllowFlashcall == false {
	//		// TODO(@benqi): 数据包非法
	//	}
	//}

	// TODO(@benqi): 独立出统一消息推送系统
	// 检查phpne是否存在，若存在是否在线决定是否通过短信发送或通过其他客户端发送
	// 透传AuthId，UserId，终端类型等
	// 检查满足条件的TransactionHash是否存在，可能的条件：
	//  1. is_deleted !=0 and now - created_at < 15 分钟

	// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumer := libphonenumber.NormalizeDigitsOnly(request.PhoneNumber)

	do := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE).SelectByPhoneAndApiIdAndHash(phoneNumer, request.ApiId, request.ApiHash)
	if do == nil {
		do = &dataobject.AuthPhoneTransactionsDO{}
		do.ApiId = request.ApiId
		do.ApiHash = request.ApiHash
		do.PhoneNumber = phoneNumer
		do.Code = "123456"
		do.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
		// TODO(@benqi): 生成一个32字节的随机字串
		do.TransactionHash = fmt.Sprintf("%20d", id.NextId())

		dao.GetAuthPhoneTransactionsDAO(dao.DB_MASTER).Insert(do)
	} else {
		// TODO(@benqi): 检查是否已经过了失效期
	}

	authSentCode := &mtproto.TLAuthSentCode{}
	authSentCode.Type = mtproto.MakeAuth_SentCodeType(&mtproto.TLAuthSentCodeTypeApp{
		Length: 6,
	})
	authSentCode.PhoneCodeHash = do.TransactionHash

	reply := mtproto.MakeAuth_SentCode(authSentCode)
	glog.Infof("AuthSendCode - reply: {%v}\n", reply)
	return reply, nil
}

func (s *AuthServiceImpl) AuthResendCode(ctx context.Context, request *mtproto.TLAuthResendCode) (*mtproto.Auth_SentCode, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthSignUp(ctx context.Context, request *mtproto.TLAuthSignUp) (*mtproto.Auth_Authorization, error) {
	glog.Infof("Process: %v", request)

	//// auth.signUp#1b067634 phone_number:string phone_code_hash:string phone_code:string first_name:string last_name:string = auth.Authorization;
	//type TLAuthSignUp struct {
	//	PhoneNumber   string `protobuf:"bytes,1,opt,name=phone_number,json=phoneNumber" json:"phone_number,omitempty"`
	//	PhoneCodeHash string `protobuf:"bytes,2,opt,name=phone_code_hash,json=phoneCodeHash" json:"phone_code_hash,omitempty"`
	//	PhoneCode     string `protobuf:"bytes,3,opt,name=phone_code,json=phoneCode" json:"phone_code,omitempty"`
	//	FirstName     string `protobuf:"bytes,4,opt,name=first_name,json=firstName" json:"first_name,omitempty"`
	//	LastName      string `protobuf:"bytes,5,opt,name=last_name,json=lastName" json:"last_name,omitempty"`
	//}

	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthSignIn(ctx context.Context, request *mtproto.TLAuthSignIn) (*mtproto.Auth_Authorization, error) {
	glog.Infof("AuthSignIn - Process: {%v}", request)

	md := grpc_util.RpcMetadataFromIncoming(ctx)
	// 客户端发送的手机号格式为: "+86 111 1111 1111"，归一化
	phoneNumer := libphonenumber.NormalizeDigitsOnly(request.PhoneNumber)

	// Check code
	authPhoneTransactionsDAO := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE)
	do1 := authPhoneTransactionsDAO.SelectByPhoneCode(request.PhoneCodeHash, request.PhoneCode, phoneNumer)
	if do1 == nil {
		err := fmt.Errorf("SelectByPhoneCode(_) return empty in request: {}%v", request)
		glog.Error(err)
		return nil, err
	}

	usersDAO := dao.GetUsersDAO(dao.DB_SLAVE)
	do2 := usersDAO.SelectByPhoneNumber(phoneNumer)
	if do2 == nil {
		if do1 == nil {
			err := fmt.Errorf("SelectByPhoneNumber(_) return empty in request{}%v", request)
			glog.Error(err)
			return nil, err
		}
	}

	do3 := dao.GetAuthUsersDAO(dao.DB_SLAVE).SelectByAuthId(md.AuthId)
	if do3 == nil {
		do3 := &dataobject.AuthUsersDO{}
		do3.AuthId = md.AuthId
		do3.UserId = do2.Id
		dao.GetAuthUsersDAO(dao.DB_MASTER).Insert(do3)
	}

	// TODO(@benqi): 从数据库加载
	authAuthorization := &mtproto.TLAuthAuthorization{}
	user := &mtproto.TLUser{}
	user.Self = true
	user.Id = do2.Id
	user.AccessHash = do2.AccessHash
	user.FirstName = do2.FirstName
	user.LastName = do2.LastName
	user.Username = do2.Username
	user.Phone = phoneNumer
	authAuthorization.User = mtproto.MakeUser(user)

	reply := mtproto.MakeAuth_Authorization(authAuthorization)
	glog.Infof("AuthSignIn - reply: {%v}\n", reply)
	return reply, nil
}

func (s *AuthServiceImpl) AuthImportAuthorization(ctx context.Context, request *mtproto.TLAuthImportAuthorization) (*mtproto.Auth_Authorization, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthImportBotAuthorization(ctx context.Context, request *mtproto.TLAuthImportBotAuthorization) (*mtproto.Auth_Authorization, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthCheckPassword(ctx context.Context, request *mtproto.TLAuthCheckPassword) (*mtproto.Auth_Authorization, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthRecoverPassword(ctx context.Context, request *mtproto.TLAuthRecoverPassword) (*mtproto.Auth_Authorization, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthExportAuthorization(ctx context.Context, request *mtproto.TLAuthExportAuthorization) (*mtproto.Auth_ExportedAuthorization, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}

func (s *AuthServiceImpl) AuthRequestPasswordRecovery(ctx context.Context, request *mtproto.TLAuthRequestPasswordRecovery) (*mtproto.Auth_PasswordRecovery, error) {
	glog.Infof("Process: %v", request)
	return nil, errors.New("Not impl")
}
