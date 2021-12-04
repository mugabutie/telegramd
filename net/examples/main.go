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
	"fmt"
	net2 "github.com/mugabutie/telegramd/net"
	"github.com/mugabutie/telegramd/net/codec"
	"log"
)

func main() {
	lengthBasedFrame := codec.NewLengthBasedFrame(1024)

	server, err := net2.Listen("tcp", "0.0.0.0:12345",
		lengthBasedFrame, 0, /* sync send */
		net2.HandlerFunc(serverSessionLoop))
	if err != nil {
		log.Fatal(err)
	}

	server.Listener().Addr().String()
	server.Serve()
}

func serverSessionLoop(session *net2.Session) {
	log.Println("OnNewSession: ")
	for {
		line, err := session.Receive()
		if err != nil {
			return
		}

		fmt.Print(line)
		err = session.Send(line)
		if err != nil {
			return
		}
	}
}
