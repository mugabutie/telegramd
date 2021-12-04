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

package main

import (
	"flag"
	"github.com/golang/glog"

	"fmt"
	"github.com/BurntSushi/toml"
	"../base/mysql_client"
	"../base/redis_client"
	"../biz_model/dal/dao"
	account "../biz_server/account/rpc"
	auth "../biz_server/auth/rpc"
	bots "../biz_server/bots/rpc"
	channels "../biz_server/channels/rpc"
	contacts "../biz_server/contacts/rpc"
	"../biz_server/delivery"
	help "../biz_server/help/rpc"
	langpack "../biz_server/langpack/rpc"
	messages "../biz_server/messages/rpc"
	payments "../biz_server/payments/rpc"
	phone "../biz_server/phone/rpc"
	photos "../biz_server/photos/rpc"
	stickers "../biz_server/stickers/rpc"
	updates "../biz_server/updates/rpc"
	upload "../biz_server/upload/rpc"
	users "../biz_server/users/rpc"
	"../grpc_util"
	"../grpc_util/middleware/recovery2"
	"../mtproto"
	"net"
)

func init() {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "false")
}

type RpcServerConfig struct {
	Addr string
}

type RpcClientConfig struct {
	ServiceName string
	Addr        string
}

type BizServerConfig struct {
	Server    *RpcServerConfig
	RpcClient *RpcClientConfig
	Mysql     []mysql_client.MySQLConfig
	Redis     []redis_client.RedisConfig
}

// 整合各服务，方便开发调试
func main() {
	flag.Parse()

	bizServerConfig := &BizServerConfig{}
	if _, err := toml.DecodeFile("./biz_server.toml", bizServerConfig); err != nil {
		fmt.Errorf("%s\n", err)
		return
	}

	glog.Info(bizServerConfig)

	// 初始化mysql_client、redis_client
	redis_client.InstallRedisClientManager(bizServerConfig.Redis)
	mysql_client.InstallMysqlClientManager(bizServerConfig.Mysql)

	// 初始化redis_dao、mysql_dao
	dao.InstallMysqlDAOManager(mysql_client.GetMysqlClientManager())
	dao.InstallRedisDAOManager(redis_client.GetRedisClientManager())

	lis, err := net.Listen("tcp", bizServerConfig.Server.Addr)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}

	delivery.InstallDeliveryInstance(bizServerConfig.RpcClient.Addr)

	// var opts []grpc.ServerOption
	// grpcServer := grpc.NewServer(opts...)
	grpcServer := grpc_recovery2.NewRecoveryServer(grpc_util.BizUnaryRecoveryHandler, grpc_util.BizStreamRecoveryHandler)

	// AccountServiceImpl
	mtproto.RegisterRPCAccountServer(grpcServer, &account.AccountServiceImpl{})

	// AuthServiceImpl
	mtproto.RegisterRPCAuthServer(grpcServer, &auth.AuthServiceImpl{})

	mtproto.RegisterRPCBotsServer(grpcServer, &bots.BotsServiceImpl{})
	mtproto.RegisterRPCChannelsServer(grpcServer, &channels.ChannelsServiceImpl{})

	// ContactsServiceImpl
	mtproto.RegisterRPCContactsServer(grpcServer, &contacts.ContactsServiceImpl{})

	mtproto.RegisterRPCHelpServer(grpcServer, &help.HelpServiceImpl{})
	mtproto.RegisterRPCLangpackServer(grpcServer, &langpack.LangpackServiceImpl{})

	// MessagesServiceImpl
	mtproto.RegisterRPCMessagesServer(grpcServer, &messages.MessagesServiceImpl{})

	mtproto.RegisterRPCPaymentsServer(grpcServer, &payments.PaymentsServiceImpl{})
	mtproto.RegisterRPCPhoneServer(grpcServer, &phone.PhoneServiceImpl{})
	mtproto.RegisterRPCPhotosServer(grpcServer, &photos.PhotosServiceImpl{})
	mtproto.RegisterRPCStickersServer(grpcServer, &stickers.StickersServiceImpl{})
	mtproto.RegisterRPCUpdatesServer(grpcServer, &updates.UpdatesServiceImpl{})
	mtproto.RegisterRPCUploadServer(grpcServer, &upload.UploadServiceImpl{})

	mtproto.RegisterRPCUsersServer(grpcServer, &users.UsersServiceImpl{})

	glog.Infof("NewRPCServer in {%v}.", bizServerConfig)

	grpcServer.Serve(lis)
}
