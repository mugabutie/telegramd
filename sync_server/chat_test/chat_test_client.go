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

package chat_test

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/mugabutie/telegramd/base/base"
	"github.com/mugabutie/telegramd/zproto"
	"google.golang.org/grpc"
	"io"
	"math/rand"
	"time"
)

func DoChatTestClient() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("TestRPCClient...")
	conn, err := grpc.Dial("127.0.0.1:22345", grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("fail to dial: %v\n", err)
	}
	defer conn.Close()
	client := zproto.NewChatTestClient(conn)

	sess := &zproto.ChatSession{base.Int64ToString(rand.Int63())}
	fmt.Println("sessionId : ", sess.SessionId)

	stream, err := client.Connect(context.Background(), &zproto.ChatSession{base.Int64ToString(rand.Int63())})
	if err != nil {
		glog.Fatalln("connect:", err)
	}

	chatMessages := make(chan *zproto.ChatMessage, 1000)
	go func() {
		defer func() { close(chatMessages) }()
		for {
			chatMessage, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				glog.Fatalln("stream.Recv", err)
			}
			chatMessages <- chatMessage
		}
	}()

	go func() {
		for {
			select {
			case chat := <-chatMessages:
				fmt.Printf("Recv chat_message: {session_id: %s, message_data: %s}\n", chat.SenderSessionId, chat.MessageData)
			}
		}
	}()

	var message string
	for {
		fmt.Print("> ")
		if n, err := fmt.Scanln(&message); err == io.EOF {
			return
		} else if n > 0 {
			_, err := client.SendChat(context.Background(), &zproto.ChatMessage{SenderSessionId: sess.SessionId, MessageData: message})
			if err != nil {
				glog.Fatalln("sendChat:", err)
			}
		}
	}
}
