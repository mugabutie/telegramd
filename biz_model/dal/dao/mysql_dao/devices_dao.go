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

package mysql_dao

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/jmoiron/sqlx"
	"../../../../biz_model/dal/dataobject"
	"../../../../mtproto"
)

type DevicesDAO struct {
	db *sqlx.DB
}

func NewDevicesDAO(db *sqlx.DB) *DevicesDAO {
	return &DevicesDAO{db}
}

// insert into devices(auth_id, user_id, token_type, token, state) values (:auth_id, :user_id, :token_type, :token, :state)
// TODO(@benqi): sqlmap
func (dao *DevicesDAO) Insert(do *dataobject.DevicesDO) int64 {
	var query = "insert into devices(auth_id, user_id, token_type, token, state) values (:auth_id, :user_id, :token_type, :token, :state)"
	r, err := dao.db.NamedExec(query, do)
	if err != nil {
		errDesc := fmt.Sprintf("NamedExec in Insert(%v), error: %v", do, err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	id, err := r.LastInsertId()
	if err != nil {
		errDesc := fmt.Sprintf("LastInsertId in Insert(%v)_error: %v", do, err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}
	return id
}

// select id from devices where auth_id = :auth_id and token_type = :token_type and token = :token limit 1
// TODO(@benqi): sqlmap
func (dao *DevicesDAO) SelectId(auth_id int64, token_type int8, token string) *dataobject.DevicesDO {
	var query = "select id from devices where auth_id = ? and token_type = ? and token = ? limit 1"
	rows, err := dao.db.Queryx(query, auth_id, token_type, token)

	if err != nil {
		errDesc := fmt.Sprintf("Queryx in SelectId(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	defer rows.Close()

	do := &dataobject.DevicesDO{}
	if rows.Next() {
		err = rows.StructScan(do)
		if err != nil {
			errDesc := fmt.Sprintf("StructScan in SelectId(_), error: %v", err)
			glog.Error(errDesc)
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
		}
	} else {
		return nil
	}

	return do
}

// update devices set state = :state where id = :id
// TODO(@benqi): sqlmap
func (dao *DevicesDAO) UpdateStateById(state int8, id int32) int64 {
	var query = "update devices set state = ? where id = ?"
	r, err := dao.db.Exec(query, state, id)

	if err != nil {
		errDesc := fmt.Sprintf("Exec in UpdateStateById(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	rows, err := r.RowsAffected()
	if err != nil {
		errDesc := fmt.Sprintf("RowsAffected in UpdateStateById(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return rows
}

// update devices set state = :state where auth_id = :auth_id and token_type = :token_type and token = :token
// TODO(@benqi): sqlmap
func (dao *DevicesDAO) UpdateState(state int8, auth_id int64, token_type int8, token string) int64 {
	var query = "update devices set state = ? where auth_id = ? and token_type = ? and token = ?"
	r, err := dao.db.Exec(query, state, auth_id, token_type, token)

	if err != nil {
		errDesc := fmt.Sprintf("Exec in UpdateState(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	rows, err := r.RowsAffected()
	if err != nil {
		errDesc := fmt.Sprintf("RowsAffected in UpdateState(_), error: %v", err)
		glog.Error(errDesc)
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_DBERR), errDesc))
	}

	return rows
}
