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

package redis_client

import (
	"github.com/mugabutie/telegramd/base/base"
)

// Redis client config.
type RedisConfig struct {
	Name         string // redis name
	Addr         string
	Active       int // pool
	Idle         int // pool
	DialTimeout  base.Duration
	ReadTimeout  base.Duration
	WriteTimeout base.Duration
	IdleTimeout  base.Duration

	DBNum    string // db号
	Password string // 密码
}
