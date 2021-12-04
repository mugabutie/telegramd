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

package model

import (
	"github.com/mugabutie/telegramd/base/base"
	"github.com/mugabutie/telegramd/biz_model/dal/dao"
	"github.com/mugabutie/telegramd/biz_model/dal/dataobject"
	"github.com/mugabutie/telegramd/frontend/id"
	"github.com/mugabutie/telegramd/mtproto"
	"sync"
	"time"
)

type chatModel struct {
	// chatDAO *dao.UserDialogsDAO
}

var (
	chatInstance     *chatModel
	chatInstanceOnce sync.Once
)

func GetChatModel() *chatModel {
	chatInstanceOnce.Do(func() {
		chatInstance = &chatModel{}
	})
	return chatInstance
}

func (m *chatModel) AddChatParticipant(chatId, chatUserId, inviterId int32, participantType int8) (participant *mtproto.ChatParticipant) {
	// uId := u.GetInputUser().GetUserId()
	chatUserDO := &dataobject.ChatUsersDO{}

	chatUserDO.ChatId = chatId
	chatUserDO.CreatedAt = base.NowFormatYMDHMS()
	chatUserDO.State = 0
	chatUserDO.InvitedAt = int32(time.Now().Unix())
	chatUserDO.InviterUserId = inviterId
	chatUserDO.JoinedAt = chatUserDO.InvitedAt
	chatUserDO.UserId = chatUserId
	chatUserDO.ParticipantType = participantType
	dao.GetChatUsersDAO(dao.DB_MASTER).Insert(chatUserDO)

	if participantType == 2 {
		participant2 := &mtproto.TLChatParticipantCreator{}
		participant2.UserId = chatUserId

		participant = participant2.ToChatParticipant()
	} else if participantType == 1 {
		participant2 := &mtproto.TLChatParticipantAdmin{}
		participant2.UserId = chatUserId
		participant2.Date = chatUserDO.InvitedAt
		participant2.InviterId = inviterId

		participant = participant2.ToChatParticipant()
	} else if participantType == 0 {
		participant2 := &mtproto.TLChatParticipant{}
		participant2.UserId = chatUserId
		participant2.Date = chatUserDO.InvitedAt
		participant2.InviterId = inviterId
		// participants.Participants = append(participants.Participants, participant.ToChatParticipant())

		participant = participant2.ToChatParticipant()
	}
	return
}

/*
	chatEmpty#9ba2d800 id:int = Chat;
	chat#d91cdd54 flags:# creator:flags.0?true kicked:flags.1?true left:flags.2?true admins_enabled:flags.3?true admin:flags.4?true deactivated:flags.5?true id:int title:string photo:ChatPhoto participants_count:int date:int version:int migrated_to:flags.6?InputChannel = Chat;
	chatForbidden#7328bdb id:int title:string = Chat;
	channel#cb44b1c flags:# creator:flags.0?true left:flags.2?true editor:flags.3?true broadcast:flags.5?true verified:flags.7?true megagroup:flags.8?true restricted:flags.9?true democracy:flags.10?true signatures:flags.11?true min:flags.12?true id:int access_hash:flags.13?long title:string username:flags.6?string photo:ChatPhoto date:int version:int restriction_reason:flags.9?string admin_rights:flags.14?ChannelAdminRights banned_rights:flags.15?ChannelBannedRights = Chat;
	channelForbidden#289da732 flags:# broadcast:flags.5?true megagroup:flags.8?true id:int access_hash:long title:string until_date:flags.16?int = Chat;
*/
func (m *chatModel) CreateChat(userId int32, title string, chatUserIdList []int32, random int64) (*mtproto.TLChat, *mtproto.TLChatParticipants) {
	chat := &mtproto.TLChat{}
	// chat.Id = int32(lastInsertId)
	chat.Title = title
	chat.Photo = mtproto.MakeChatPhoto(&mtproto.TLChatPhotoEmpty{})
	chat.Date = int32(time.Now().Unix())
	chat.Version = 1
	chat.ParticipantsCount = int32(len(chatUserIdList)) + 1

	chatDO := &dataobject.ChatsDO{}
	chatDO.AccessHash = id.NextId()
	chatDO.CreatedAt = base.NowFormatYMDHMS()
	chatDO.CreatorUserId = userId
	// TODO(@benqi): 使用客户端message_id
	chatDO.CreateRandomId = id.NextId()
	chatDO.Title = title

	chatDO.TitleChangerUserId = userId
	chatDO.TitleChangedAt = chatDO.CreatedAt
	// TODO(@benqi): 使用客户端message_id
	chatDO.TitleChangeRandomId = chatDO.AccessHash

	chatDO.AvatarChangerUserId = userId
	chatDO.AvatarChangedAt = chatDO.CreatedAt
	// TODO(@benqi): 使用客户端message_id
	chatDO.AvatarChangeRandomId = chatDO.AccessHash
	// dao.GetChatsDA()
	chatDO.ParticipantCount = chat.ParticipantsCount

	// TODO(@benqi): 事务！
	chat.Id = int32(dao.GetChatsDAO(dao.DB_MASTER).Insert(chatDO))

	// updateChatParticipants := &mtproto.TLUpdateChatParticipants{}
	participants := &mtproto.TLChatParticipants{}
	participants.ChatId = chat.Id
	participants.Version = 1

	participants.Participants = append(participants.Participants, m.AddChatParticipant(chat.Id, userId, userId, 2))
	// chatUserIdList := make([]int32, 0, len(request.GetUsers()))
	for _, chatUserId := range chatUserIdList {
		if chatUserId == userId {
			continue
		}
		participants.Participants = append(participants.Participants, m.AddChatParticipant(chat.Id, chatUserId, userId, 0))
	}

	return chat, participants
}

func (m *chatModel) GetChat(chatId int32) *mtproto.TLChat {
	chat := &mtproto.TLChat{}
	chatDO := dao.GetChatsDAO(dao.DB_SLAVE).Select(chatId)
	if chatDO == nil {
		panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "InputPeer invalid"))
	}
	chat.Id = chatId
	chat.Title = chatDO.Title
	chat.Photo = mtproto.MakeChatPhoto(&mtproto.TLChatPhotoEmpty{})
	chat.Version = chatDO.Version
	chat.ParticipantsCount = chatDO.ParticipantCount
	chat.Date = int32(time.Now().Unix())
	return chat
}

func (m *chatModel) GetChatFull(chatId int32) *mtproto.TLChatFull {
	chatFull := &mtproto.TLChatFull{}

	chatFull.Id = chatId
	chatFull.Participants = m.GetChatParticipants(chatId).ToChatParticipants()
	photo := &mtproto.TLPhotoEmpty{}
	chatFull.ChatPhoto = photo.ToPhoto()
	chatFull.ExportedInvite = mtproto.MakeExportedChatInvite(&mtproto.TLChatInviteEmpty{})
	return chatFull
}

//func (m* chatModel) GetChatAndParticipants(chatId int32) (*mtproto.TLChat, *mtproto.TLChatParticipants) {
//	chat := m.GetChat(chatId)
//	participants := m.GetChatParticipants(chatId)
//	return  chat, participants
//}

func (m *chatModel) GetChatParticipants(chatId int32) *mtproto.TLChatParticipants {
	chatUsersDOList := dao.GetChatUsersDAO(dao.DB_SLAVE).SelectByChatId(chatId)

	// updateChatParticipants := &mtproto.TLUpdateChatParticipants{}
	participants := &mtproto.TLChatParticipants{}
	participants.ChatId = chatId
	participants.Version = 1
	for _, chatUsersDO := range chatUsersDOList {
		// uId := u.GetInputUser().GetUserId()
		if chatUsersDO.ParticipantType == 2 {
			// chatUserDO.IsAdmin = 1
			participant := &mtproto.TLChatParticipantCreator{}
			participant.UserId = chatUsersDO.UserId
			participants.Participants = append(participants.Participants, participant.ToChatParticipant())
		} else if chatUsersDO.ParticipantType == 1 {
			participant := &mtproto.TLChatParticipantAdmin{}
			participant.UserId = chatUsersDO.UserId
			participant.InviterId = chatUsersDO.InviterUserId
			participant.Date = chatUsersDO.JoinedAt
			participants.Participants = append(participants.Participants, participant.ToChatParticipant())
		} else if chatUsersDO.ParticipantType == 0 {
			participant := &mtproto.TLChatParticipant{}
			participant.UserId = chatUsersDO.UserId
			participant.InviterId = chatUsersDO.InviterUserId
			participant.Date = chatUsersDO.JoinedAt
			participants.Participants = append(participants.Participants, participant.ToChatParticipant())
		}
	}
	return participants
}

func (m *chatModel) GetChatsByIDList(idList []int32) (chats []*mtproto.TLChat) {
	// TODO(@benqi): Check messageDAO
	chatsDOList := dao.GetChatsDAO(dao.DB_SLAVE).SelectByIdList(idList)

	for _, chatDO := range chatsDOList {
		chat := &mtproto.TLChat{}
		chat.Id = chatDO.Id
		chat.Title = chatDO.Title
		chat.Photo = mtproto.MakeChatPhoto(&mtproto.TLChatPhotoEmpty{})
		chat.Version = chatDO.Version
		chat.Date = int32(time.Now().Unix())
		chats = append(chats, chat)
	}
	return
}

func (m *chatModel) GetChatListByIDList(idList []int32) (chats []*mtproto.Chat) {
	// TODO(@benqi): Check messageDAO
	chatsDOList := dao.GetChatsDAO(dao.DB_SLAVE).SelectByIdList(idList)

	for _, chatDO := range chatsDOList {
		chat := &mtproto.TLChat{}
		chat.Id = chatDO.Id
		chat.Title = chatDO.Title
		chat.Photo = mtproto.MakeChatPhoto(&mtproto.TLChatPhotoEmpty{})
		chat.Version = chatDO.Version
		chat.Date = int32(time.Now().Unix())
		chats = append(chats, chat.ToChat())
	}
	return
}
